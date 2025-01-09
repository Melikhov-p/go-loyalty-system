package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type OrderRepo struct {
	logger *zap.Logger
	cfg    *config.Config
	db     *sql.DB
}

func (or *OrderRepo) CreateOrder(ctx context.Context, orderNumber string, user *models.User) error {
	query := `INSERT INTO "order" (number, uploaded_at, user_id) VALUES ($1, $2, $3)`

	_, err := or.db.ExecContext(ctx, query, orderNumber, time.Now().Format(time.DateTime), user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			var order *models.Order
			order, err = or.GetOrderByNumber(ctx, orderNumber)
			if err != nil {
				return fmt.Errorf(
					"order number maybe exist, but got error with geting order with number %s: %w",
					orderNumber,
					err)
			}

			if order.UserID == user.ID {
				return ErrOrderByUserExist
			}
			return ErrOrderNumberExist
		}
		return fmt.Errorf("error executing context for create order %w", err)
	}

	return nil
}

func (or *OrderRepo) CreateWatchedOrder(ctx context.Context, orderNumber string, user *models.User) error {
	query := `INSERT INTO watched_order (order_number, user_id) VALUES ($1, $2)`

	_, err := or.db.ExecContext(ctx, query, orderNumber, user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrOrderNumberExist
		}
		return fmt.Errorf("error executing context for create watched order %w", err)
	}

	return nil
}

func (or *OrderRepo) GetOrderByNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	query := `SELECT id, status, accrual, uploaded_at, user_id FROM "order" WHERE number=$1`

	order := NewEmptyOrder()
	order.Number = orderNumber

	row := or.db.QueryRowContext(ctx, query, orderNumber)

	if err := row.Scan(&order.ID, &order.Status, &order.Accrual, &order.UploadedAt, &order.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNumberNotFound
		}
		return nil, fmt.Errorf("error scannig row for order with number %s: %w", orderNumber, err)
	}

	return order, nil
}

func (or *OrderRepo) GetOrdersByUser(ctx context.Context, userID int) ([]*models.Order, error) {
	query := `SELECT id, number, status, accrual, uploaded_at FROM "order" WHERE user_id=$1`

	rows, err := or.db.QueryContext(ctx, query, userID)
	defer func() {
		_ = rows.Close()
	}()

	if err != nil {
		return nil, fmt.Errorf("error query for orders by userID %d: %w", userID, err)
	}

	var orders []*models.Order
	for rows.Next() {
		order := NewEmptyOrder()
		order.UserID = userID
		if err = rows.Scan(&order.ID, &order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, fmt.Errorf("error scanning row of order %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("got rows.Err() : %w", err)
	}

	if len(orders) == 0 {
		return nil, ErrOrdersNotFound
	}
	return orders, nil
}

func (or *OrderRepo) GetWatchedOrders(ctx context.Context) ([]*models.WatchedOrder, error) {
	query := `SELECT id, order_number, user_id, accrual_order_status FROM watched_order WHERE trackable = true`

	rows, err := or.db.QueryContext(ctx, query)
	defer func() {
		_ = rows.Close()
	}()
	if err != nil {
		return nil, fmt.Errorf("error query context for getting watched orders %w", err)
	}

	var watchedOrders []*models.WatchedOrder
	for rows.Next() {
		var order models.WatchedOrder
		if err = rows.Scan(&order.ID, &order.OrderNumber, &order.UserID, &order.AccrualOrderStatus); err != nil {
			return nil, fmt.Errorf("error scanning row for watched order %w", err)
		}

		watchedOrders = append(watchedOrders, &order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("got rows.Err(): %w", err)
	}

	return watchedOrders, nil
}

func (or *OrderRepo) UpdateOrdersStatus(ctx context.Context, orders []*models.WatchedOrder) error {
	query := `UPDATE "order" SET status = $1, accrual = $2 WHERE number = $3`

	tx, err := or.db.Begin()
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()
	if err != nil {
		return fmt.Errorf("error starting transaction for update order status %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, query)
	defer func() {
		_ = stmt.Close()
	}()
	if err != nil {
		return fmt.Errorf("error preparing context for update order status query %w", err)
	}

	for _, order := range orders {
		_, err = stmt.ExecContext(ctx, order.AccrualOrderStatus, order.AccrualPoints, order.OrderNumber)
		if err != nil {
			return fmt.Errorf("error exectunig context for update order status %w", err)
		}
	}

	return nil
}

func (or *OrderRepo) StopWatchOrder(ctx context.Context, order *models.WatchedOrder) error {
	query := `UPDaTE watched_order SET trackable = false WHERE order_number = $1`

	_, err := or.db.ExecContext(ctx, query, order.OrderNumber)
	if err != nil {
		return fmt.Errorf("error stoping watch order %s: %w", order.OrderNumber, err)
	}

	return nil
}

func NewEmptyOrder() *models.Order {
	return &models.Order{
		ID:         -1,
		Number:     "",
		Status:     "NEW",
		Accrual:    nil,
		UploadedAt: time.Now(),
		UserID:     -1,
	}
}

func NewOrderRepo(logger *zap.Logger, cfg *config.Config, db *sql.DB) *OrderRepo {
	return &OrderRepo{
		logger: logger,
		cfg:    cfg,
		db:     db,
	}
}
