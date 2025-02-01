package workers

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/Melikhov-p/go-loyalty-system/internal/services"
	"github.com/Melikhov-p/go-loyalty-system/internal/workers/worker"
	"go.uber.org/zap"
)

type Worker interface {
	Run()
	Stop(wg *sync.WaitGroup)
	GetID() int
}

type Dispatcher struct {
	log               *zap.Logger
	cfg               *config.Config
	db                *sql.DB
	pingInterval      time.Duration
	sleepPoint        time.Time
	workersSleepPoint time.Time
	orderService      *services.OrderService
	maxWorkers        int
	workers           []Worker
	stopCh            chan interface{}
	tasks             chan *models.WatchedOrder
	ordersToUpdate    chan *models.WatchedOrder
	once              sync.Once
}

func NewDispatcher(log *zap.Logger,
	maxWorkers int,
	cfg *config.Config,
	db *sql.DB,
	pingInterval time.Duration,
) *Dispatcher {
	return &Dispatcher{
		log:               log,
		cfg:               cfg,
		db:                db,
		pingInterval:      pingInterval,
		sleepPoint:        time.Time{},
		workersSleepPoint: time.Time{},
		orderService:      services.NewOrderService(log, cfg, db),
		maxWorkers:        maxWorkers,
		workers:           make([]Worker, 0, maxWorkers),
		stopCh:            make(chan interface{}),
		tasks:             make(chan *models.WatchedOrder, maxWorkers),
		ordersToUpdate:    make(chan *models.WatchedOrder, maxWorkers),
		once:              sync.Once{},
	}
}

func (d *Dispatcher) Run() {
	var once sync.Once
	once.Do(d.HireWorkers)

	for _, w := range d.workers {
		w := w
		go func() {
			w.Run()
		}()
	}

	for {
		if d.sleepPoint.After(time.Now()) {
			continue
		}
		d.sleepPoint = time.Time{}
		d.UnRestWorkers()
		select {
		case <-d.stopCh:
			close(d.stopCh)
			wg := sync.WaitGroup{}

			for _, w := range d.workers {
				w := w // избегаем замыкания типа
				wg.Add(1)
				w.Stop(&wg)
			}
			wg.Wait()

			d.log.Debug("dispatcher has been shutdown")
			return
		case fOrder := <-d.ordersToUpdate:
			orders := make([]*models.WatchedOrder, 0, d.maxWorkers)
			orders = append(orders, fOrder)
			d.log.Debug("dispatcher case orderToUpdate channel", zap.Int("orders_count", len(d.ordersToUpdate)))

			for len(d.ordersToUpdate) > 0 {
				order := <-d.ordersToUpdate
				d.log.Debug("dispatcher find order to update", zap.Int("ID", order.ID))
				orders = append(orders, order)
			}

			err := d.UpdateOrderStatus(orders)
			if err != nil {
				d.log.Error("error updating orders", zap.Error(err))
			}
			d.sleepPoint = time.Now().Add(d.pingInterval)
		default:
			d.log.Debug("dispatcher check new tasks for update")
			err := d.CheckNewTasks()
			if err != nil {
				d.log.Error("error checking new tasks for update", zap.Error(err))
			}
			d.sleepPoint = time.Now().Add(d.pingInterval)
		}
	}
}

func (d *Dispatcher) Stop() {
	d.once.Do(func() {
		d.stopCh <- 1
	})
}

func (d *Dispatcher) HireWorkers() {
	for i := range d.maxWorkers {
		newWorker := worker.NewWorker(d.log, d.tasks, d.ordersToUpdate, d, d.cfg, i)
		d.log.Debug("dispatcher got new worker", zap.Int("ID", newWorker.GetID()), zap.Any("Worker", newWorker))
		d.workers = append(d.workers, newWorker)
	}
}

func (d *Dispatcher) IsWorkTime() bool {
	if d.workersSleepPoint.IsZero() {
		return true
	}

	if time.Now().After(d.workersSleepPoint) {
		d.UnRestWorkers()
		return true
	}

	return false
}

func (d *Dispatcher) RestWorkers(period time.Duration) {
	d.workersSleepPoint = time.Now().Add(period)
}

func (d *Dispatcher) UnRestWorkers() {
	d.workersSleepPoint = time.Time{}
}

func (d *Dispatcher) CheckNewTasks() error {
	watchedOrders, err := d.orderService.GetWatchedOrders(context.Background())
	if err != nil {
		d.log.Error("Order Watcher: error getting watched orders", zap.Error(err))
		return fmt.Errorf("error get watched orders %w", err)
	}

	for _, task := range watchedOrders {
		d.tasks <- task
		d.log.Debug("add new task", zap.Int("ID", task.ID))
	}

	return nil
}

func (d *Dispatcher) UpdateOrderStatus(orders []*models.WatchedOrder) error {
	err := d.orderService.UpdateOrderStatus(context.Background(), orders)
	if err != nil {
		return fmt.Errorf("error updating status for orders: %w", err)
	}

	return nil
}

func (d *Dispatcher) FireWorker(w Worker) {
	for i, v := range d.workers {
		if v == w {
			// Удаляем воркера по индексу
			d.workers = append(d.workers[:i], d.workers[i+1:]...)
		}
	} // Если элемент не найден, воркеры остаются те же
}
