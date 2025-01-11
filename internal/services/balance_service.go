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

type BalanceService struct {
	logger      *zap.Logger
	cfg         *config.Config
	BalanceRepo *repository.BalanceRepo
}

func NewBalanceService(logger *zap.Logger, cfg *config.Config, db *sql.DB) *BalanceService {
	return &BalanceService{
		logger:      logger,
		cfg:         cfg,
		BalanceRepo: repository.NewBalanceRepo(logger, cfg, db),
	}
}

func (bs *BalanceService) GetUserBalance(ctx context.Context, user *models.User) error {
	err := bs.BalanceRepo.GetUserBalance(ctx, user)
	if err != nil {
		return fmt.Errorf("error getting user balance %w", err)
	}

	return nil
}

func (bs *BalanceService) AddNewBalanceForUser(ctx context.Context, user *models.User) error {
	err := bs.BalanceRepo.AddNewBalanceForUser(ctx, user)
	if err != nil {
		return fmt.Errorf("error adding new balance for user %w", err)
	}

	return nil
}

func (bs *BalanceService) Withdraw(
	ctx context.Context,
	order *models.Order,
	user *models.User,
	sum float64,
) (*models.Balance, error) {
	if user.BalanceInfo.Current < sum {
		return nil, ErrNotEnough
	}

	ctx, cancel := context.WithTimeout(ctx, bs.cfg.DB.ContextTimeout)
	defer cancel()

	newBalance, err := bs.BalanceRepo.Withdraw(ctx, order, user, sum)
	if err != nil {
		return nil, fmt.Errorf("error withdraw %w", err)
	}

	return newBalance, nil
}

func (bs *BalanceService) GetUserWithdrawHistory(
	ctx context.Context,
	user *models.User,
) ([]*models.WithdrawHistoryItem, error) {
	ctx, cancel := context.WithTimeout(ctx, bs.cfg.DB.ContextTimeout)
	defer cancel()

	history, err := bs.BalanceRepo.GetUserHistory(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error getting user balance history %w", err)
	}

	return history, nil
}

func (bs *BalanceService) IncreaseBalance(ctx context.Context, userID int, diff float64) error {
	err := bs.BalanceRepo.IncreaseBalanceByUserID(ctx, userID, diff)
	if err != nil {
		return fmt.Errorf("error increasing user balance %w", err)
	}

	return nil
}
