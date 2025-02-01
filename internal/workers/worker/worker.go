package worker

import (
	"errors"
	"sync"
	"time"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/Melikhov-p/go-loyalty-system/internal/services"
	"go.uber.org/zap"
)

type Dispatcher interface {
	Run()
	Stop()
	HireWorkers()
	RestWorkers(time.Duration)
	IsWorkTime() bool
}

type Worker struct {
	id             int
	log            *zap.Logger
	stopCh         chan *sync.WaitGroup
	accrualService *services.AccrualService
	dispatcher     Dispatcher
	taskCh         chan *models.WatchedOrder
	ordersToUpdate chan *models.WatchedOrder
	once           sync.Once
}

func NewWorker(log *zap.Logger,
	taskCh chan *models.WatchedOrder,
	ordersToUpdate chan *models.WatchedOrder,
	dispatcher Dispatcher,
	cfg *config.Config,
	id int) *Worker {
	return &Worker{
		id:             id,
		log:            log,
		stopCh:         make(chan *sync.WaitGroup),
		accrualService: services.NewAccrualService(log, cfg),
		taskCh:         taskCh,
		dispatcher:     dispatcher,
		ordersToUpdate: ordersToUpdate,
		once:           sync.Once{},
	}
}

func (w *Worker) GetID() int {
	return w.id
}

func (w *Worker) Run() {
	for {
		if !w.dispatcher.IsWorkTime() {
			continue
		}
		select {
		case wg := <-w.stopCh:
			close(w.stopCh)
			w.log.Debug("worker stopped", zap.Int("WorkerID", w.id))
			wg.Done()
			return
		case task := <-w.taskCh:
			w.log.Debug("worker found new task", zap.Int("WorkerID", w.id))
			orderToUpdate, retryAfter, err := w.accrualService.CheckOrdersStatus(task)
			if err != nil {
				if errors.Is(err, services.ErrRetryAfter) {
					w.log.Debug("got request limit error")
					w.dispatcher.RestWorkers(retryAfter)
					continue
				} else {
					w.log.Error("error checking order accrual status in accrual service",
						zap.Error(err),
						zap.Int("OrderID", task.ID))
				}
			}

			w.ordersToUpdate <- orderToUpdate
			w.log.Debug("worker add order to update",
				zap.Int("ID", orderToUpdate.ID),
				zap.Int("WorkerID", w.id))
		}
	}
}

func (w *Worker) Stop(wg *sync.WaitGroup) {
	w.stopCh <- wg
}
