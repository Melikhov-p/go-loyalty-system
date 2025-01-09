package handlers

import (
	"database/sql"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/services"
	"go.uber.org/zap"
)

type Handlers struct {
	ForUser    *UserHandlers
	ForBalance *BalanceHandlers
	ForOrder   *OrderHandlers
}

type UserHandlers struct {
	logger      *zap.Logger
	cfg         *config.Config
	userService *services.UserService
}

type BalanceHandlers struct {
	logger         *zap.Logger
	cfg            *config.Config
	balanceService *services.BalanceService
	orderService   *services.OrderService
}

type OrderHandlers struct {
	logger       *zap.Logger
	cfg          *config.Config
	orderService *services.OrderService
}

func SetupHandlers(logger *zap.Logger, cfg *config.Config, db *sql.DB) *Handlers {
	return &Handlers{
		ForUser: &UserHandlers{
			logger:      logger,
			cfg:         cfg,
			userService: services.NewUserService(logger, cfg, db),
		},
		ForBalance: &BalanceHandlers{
			logger:         logger,
			cfg:            cfg,
			balanceService: services.NewBalanceService(logger, cfg, db),
			orderService:   services.NewOrderService(logger, cfg, db),
		},
		ForOrder: &OrderHandlers{
			logger:       logger,
			cfg:          cfg,
			orderService: services.NewOrderService(logger, cfg, db),
		},
	}
}
