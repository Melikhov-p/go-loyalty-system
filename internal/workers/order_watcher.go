package workers

import (
	"context"
	"database/sql"
	"time"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/services"
	"go.uber.org/zap"
)

type OrderWatcher struct {
	db             *sql.DB
	logger         *zap.Logger
	cfg            *config.Config
	stopCh         chan interface{}
	pingInterval   time.Duration
	orderService   *services.OrderService
	accrualService *services.AccrualService
}

func NewOrderWatcher(
	db *sql.DB,
	logger *zap.Logger,
	cfg *config.Config,
	pingInterval time.Duration,
) *OrderWatcher {
	return &OrderWatcher{
		db:             db,
		logger:         logger,
		cfg:            cfg,
		stopCh:         make(chan interface{}),
		pingInterval:   pingInterval,
		orderService:   services.NewOrderService(logger, cfg, db),
		accrualService: services.NewAccrualService(logger, cfg),
	}
}

func (ow *OrderWatcher) Work() {
	for {
		select {
		case <-ow.stopCh:
			ow.logger.Debug("OrderWatcher stop working")
			close(ow.stopCh)
			return
		default:
			watchedOrders, err := ow.orderService.GetWatchedOrders(context.Background())
			if err != nil {
				ow.logger.Error("Order Watcher: error getting watched orders", zap.Error(err))
				time.Sleep(ow.pingInterval)
				continue
			}

			updatedOrders, err := ow.accrualService.CheckOrdersStatus(watchedOrders)
			if err != nil {
				ow.logger.Error("Order Watcher: error checking orders statuses", zap.Error(err))
				time.Sleep(ow.pingInterval)
				continue
			}

			if len(updatedOrders) != 0 {
				ow.logger.Debug("got orders to update", zap.Any("ORDERS", updatedOrders))
				err = ow.orderService.UpdateOrdersStatus(context.Background(), updatedOrders)
				if err != nil {
					ow.logger.Error("error updating orders statuses",
						zap.Any("orders", updatedOrders),
						zap.Error(err))
					time.Sleep(ow.pingInterval)
					continue
				}
			}

			time.Sleep(ow.pingInterval)
		}
	}
}

func (ow *OrderWatcher) Stop() {
	ow.stopCh <- 1
	return
}
