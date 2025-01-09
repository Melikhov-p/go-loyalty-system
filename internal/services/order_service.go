package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/Melikhov-p/go-loyalty-system/internal/repository"
	"go.uber.org/zap"
)

type OrderService struct {
	logger    *zap.Logger
	cfg       *config.Config
	OrderRepo *repository.OrderRepo
}

func NewOrderService(logger *zap.Logger, cfg *config.Config, db *sql.DB) *OrderService {
	return &OrderService{
		logger:    logger,
		cfg:       cfg,
		OrderRepo: repository.NewOrderRepo(logger, cfg, db),
	}
}

func (os *OrderService) CreateOrder(ctx context.Context, orderNumber string, user *models.User) error {
	ctx, cancel := context.WithTimeout(ctx, os.cfg.DB.ContextTimeout)
	defer cancel()

	err := os.OrderRepo.CreateOrder(ctx, orderNumber, user)
	if err != nil {
		return fmt.Errorf("error creating new order %w", err)
	}

	err = os.OrderRepo.CreateWatchedOrder(ctx, orderNumber, user)
	if err != nil {
		return fmt.Errorf("error creating new watched order %w", err)
	}

	return nil
}

func (os *OrderService) GetOrdersByUser(ctx context.Context, user *models.User) ([]*models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, os.cfg.DB.ContextTimeout)
	defer cancel()

	orders, err := os.OrderRepo.GetOrdersByUser(ctx, user.ID)
	if err != nil {
		return []*models.Order{}, fmt.Errorf("error getting orders by user with id %d: %w", user.ID, err)
	}
	return orders, nil
}

func (os *OrderService) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, os.cfg.DB.ContextTimeout)
	defer cancel()

	order, err := os.OrderRepo.GetOrderByNumber(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("error getting order by number %s: %w", number, err)
	}

	return order, nil
}

func (os *OrderService) GetWatchedOrders(ctx context.Context) ([]*models.WatchedOrder, error) {
	ctx, cancel := context.WithTimeout(ctx, os.cfg.DB.ContextTimeout)
	defer cancel()

	orders, err := os.OrderRepo.GetWatchedOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting watched orders %w", err)
	}

	return orders, nil
}

func (os *OrderService) UpdateOrdersStatus(ctx context.Context, orders []*models.WatchedOrder) error {
	ctx, cancel := context.WithTimeout(ctx, os.cfg.DB.ContextTimeout)
	defer cancel()

	err := os.OrderRepo.UpdateOrdersStatus(ctx, orders)
	if err != nil {
		return fmt.Errorf("error updating orders status %w", err)
	}

	return nil
}