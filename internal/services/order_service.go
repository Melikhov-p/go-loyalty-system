package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/Melikhov-p/go-loyalty-system/internal/repository"
	"go.uber.org/zap"
)

type OrderFinalStatus string

const (
	invalid   OrderFinalStatus = "INVALID"
	processed OrderFinalStatus = "PROCESSED"
)

type OrderService struct {
	logger         *zap.Logger
	cfg            *config.Config
	OrderRepo      *repository.OrderRepo
	BalanceService *BalanceService
}

func NewOrderService(logger *zap.Logger, cfg *config.Config, db *sql.DB) *OrderService {
	return &OrderService{
		logger:         logger,
		cfg:            cfg,
		OrderRepo:      repository.NewOrderRepo(logger, cfg, db),
		BalanceService: NewBalanceService(logger, cfg, db),
	}
}

// ValidateOrderNumber Функция для проверки номера заказа с использованием алгоритма Луна.
func (os *OrderService) ValidateOrderNumber(orderNumber string) bool {
	// Удаляем все нецифровые символы из номера заказа
	var cleanNumber string
	for _, char := range orderNumber {
		if char >= '0' && char <= '9' {
			cleanNumber += string(char)
		}
	}

	// Проверяем, что номер заказа содержит только цифры
	if len(cleanNumber) == 0 {
		return false
	}

	sum := 0
	reverse := false

	// Итерируемся по цифрам номера заказа справа налево
	for i := len(cleanNumber) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(cleanNumber[i]))
		if reverse {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		reverse = !reverse
	}

	// Номер заказа валиден, если сумма делится на 10 без остатка
	return sum%10 == 0
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

	for _, order := range orders {
		if order.AccrualOrderStatus == string(invalid) || order.AccrualOrderStatus == string(processed) {
			os.logger.Debug("order in final status",
				zap.String("ORDER NUMBER", order.OrderNumber),
				zap.String("ORDER FINAL STATUS", order.AccrualOrderStatus))

			err = os.OrderRepo.StopWatchOrder(ctx, order)
			if err != nil {
				os.logger.Error("error stopping watch order",
					zap.String("NUMBER", order.OrderNumber),
					zap.Error(err))
			}

			err = os.BalanceService.IncreaseBalance(ctx, order.UserID, order.AccrualPoints)
			if err != nil {
				os.logger.Error("error increasing user balance",
					zap.Int("USERID", order.UserID),
					zap.Float64("DIFF", order.AccrualPoints),
					zap.Error(err),
				)
			}
		}
	}

	return nil
}
