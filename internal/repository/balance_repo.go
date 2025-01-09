package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"go.uber.org/zap"
)

type BalanceRepo struct {
	logger *zap.Logger
	cfg    *config.Config
	db     *sql.DB
}

func NewEmptyBalance() *models.Balance {
	return &models.Balance{
		Current:   -1,
		Withdrawn: -1,
	}
}

func (br *BalanceRepo) GetUserBalance(ctx context.Context, user *models.User) error {
	query := `SELECT current, withdrawn FROM balance WHERE user_id=$1`

	ctx, cancel := context.WithTimeout(ctx, br.cfg.DB.ContextTimeout)
	defer cancel()

	row := br.db.QueryRowContext(ctx, query, user.ID)
	if err := row.Scan(&user.BalanceInfo.Current, &user.BalanceInfo.Withdrawn); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = br.AddNewBalanceForUser(ctx, user)
			if err != nil {
				return fmt.Errorf("error creating new balance for user without it %w", err)
			}
			return nil
		}
		return fmt.Errorf("error scanning row for user balance %w", err)
	}

	return nil
}

func (br *BalanceRepo) AddNewBalanceForUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO balance (user_id) VALUES ($1)`

	ctx, cancel := context.WithTimeout(ctx, br.cfg.DB.ContextTimeout)
	defer cancel()

	_, err := br.db.ExecContext(ctx, query, user.ID)
	if err != nil {
		return fmt.Errorf("error executing context for create balance for user %w", err)
	}

	user.BalanceInfo.Current, user.BalanceInfo.Withdrawn = 0, 0
	return nil
}

func (br *BalanceRepo) Withdraw(
	ctx context.Context,
	order *models.Order,
	user *models.User,
	sum float64,
) (*models.Balance, error) {
	tx, err := br.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("error starting transaction for withdraw %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	withdrawQuery := `UPDATE balance
	SET
		current = current - $1,
		withdrawn = withdrawn + $1
	WHERE
		user_id = $2`

	_, err = tx.ExecContext(ctx, withdrawQuery, sum, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error executing context for withdraw query %w", err)
	}

	historyQuery := `INSERT into withdraw_history (order_number, sum, processed_at, user_id) VALUES ($1, $2, $3, $4)`

	_, err = tx.ExecContext(
		ctx,
		historyQuery,
		order.Number,
		sum,
		time.Now().Format("2006-01-02 15:04:05"),
		user.ID)
	if err != nil {
		return nil, fmt.Errorf("error executing context for history query %w", err)
	}

	newBalance := &models.Balance{
		Current:   user.BalanceInfo.Current - sum,
		Withdrawn: user.BalanceInfo.Withdrawn + sum,
	}

	return newBalance, nil
}

func (br *BalanceRepo) GetUserHistory(ctx context.Context, user *models.User) ([]*models.WithdrawHistoryItem, error) {
	query := `SELECT order_number, sum, processed_at FROM withdraw_history WHERE user_id = $1`

	rows, err := br.db.QueryContext(ctx, query, user.ID)
	if err != nil {
		return nil, fmt.Errorf("error query context for user balance history %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var history []*models.WithdrawHistoryItem
	for rows.Next() {
		var item models.WithdrawHistoryItem
		if err = rows.Scan(&item.Order, &item.Sum, &item.ProcessedAt); err != nil {
			return nil, fmt.Errorf("error scannning row for balance history %w", err)
		}
		history = append(history, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("got rows.Err() %w", err)
	}

	if len(history) == 0 {
		return nil, ErrEmptyBalanceHistory
	}

	return history, err
}

func NewBalanceRepo(logger *zap.Logger, cfg *config.Config, db *sql.DB) *BalanceRepo {
	return &BalanceRepo{
		logger: logger,
		cfg:    cfg,
		db:     db,
	}
}
