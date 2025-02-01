package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Melikhov-p/go-loyalty-system/internal/auth"
	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/models"
	"github.com/Melikhov-p/go-loyalty-system/internal/repository"
	"go.uber.org/zap"
)

type UserService struct {
	logger         *zap.Logger
	cfg            *config.Config
	UserRepo       *repository.UserRepo
	BalanceService *BalanceService
}

var ErrIncorrectPass error = errors.New("password diff")

func (us *UserService) GetUserByToken(ctx context.Context, tokenString string) (*models.User, error) {
	userID, err := auth.GetUserIDbyToken(tokenString, us.cfg.DB.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("error getting user id from token claims %w", err)
	}

	user, err := us.UserRepo.GetUserWithID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	user.AuthInfo.Token = tokenString
	err = us.BalanceService.GetUserBalance(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error getting balance for user with id %d: %w", user.ID, err)
	}

	return user, nil
}

func (us *UserService) AddNewUser(
	ctx context.Context,
	login string, password string,
) (*models.User, error) {
	passHash := auth.HashFor(password)

	user, err := us.UserRepo.AddUser(ctx, login, passHash)
	if err != nil {
		return nil, fmt.Errorf("error adding new user %w", err)
	}

	user.Password = passHash
	if user.AuthInfo.Token, err = auth.BuildJWTToken(user.ID, us.cfg.DB.SecretKey, us.cfg.TokenLifeTime); err != nil {
		err = us.UserRepo.DeleteUserWithID(ctx, user.ID)
		if err != nil {
			us.logger.Error(
				"error deleting user in case of error of building JWT token",
				zap.Int("USER_ID", user.ID),
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("error build JWT token for user %w", err)
	}

	err = us.BalanceService.AddNewBalanceForUser(ctx, user)
	if err != nil {
		err = us.UserRepo.DeleteUserWithID(ctx, user.ID)
		if err != nil {
			us.logger.Error(
				"error deleting user in case of error of building JWT token",
				zap.Int("USER_ID", user.ID),
				zap.Error(err),
			)
		}
		return nil, fmt.Errorf("error creating balance for user %w", err)
	}

	_ = us.BalanceService.GetUserBalance(ctx, user)

	return user, nil
}

func (us *UserService) AuthUser(
	ctx context.Context,
	login string,
	password string,
) (*models.User, error) {
	passHash := auth.HashFor(password)

	user, err := us.UserRepo.GetUserWithLogin(ctx, login)
	if err != nil {
		return us.UserRepo.NewEmptyUser(), fmt.Errorf("error getting user by login %w", err)
	}

	if user.Password != passHash {
		return us.UserRepo.NewEmptyUser(), ErrIncorrectPass
	}

	if user.AuthInfo.Token, err = auth.BuildJWTToken(user.ID, us.cfg.DB.SecretKey, us.cfg.TokenLifeTime); err != nil {
		return nil, fmt.Errorf("error build JWT token for user %w", err)
	}

	return user, nil
}

func NewUserService(logger *zap.Logger, cfg *config.Config, db *sql.DB) *UserService {
	return &UserService{
		logger:         logger,
		cfg:            cfg,
		UserRepo:       repository.NewUserRepo(logger, cfg, db),
		BalanceService: NewBalanceService(logger, cfg, db),
	}
}
