package router

import (
	"database/sql"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	handlersPkg "github.com/Melikhov-p/go-loyalty-system/internal/handlers"
	"github.com/Melikhov-p/go-loyalty-system/internal/middlewares"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func CreateRouter(cfg *config.Config, logger *zap.Logger, db *sql.DB) chi.Router {
	r := chi.NewRouter()

	mdlwr := middlewares.NewMiddleware(logger, cfg, db)
	r.Use(
		mdlwr.WithLogging,
		mdlwr.WithAuth,
		mdlwr.GzipMiddleware,
	)

	handlers := handlersPkg.SetupHandlers(logger, cfg, db)

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", handlers.ForUser.UserRegister)
			r.Post("/login", handlers.ForUser.UserLogin)
			r.Route("/orders", func(r chi.Router) {
				r.Post("/", handlers.ForOrder.CreateOrder)
				r.Get("/", handlers.ForOrder.GetOrders)
			})
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", handlers.ForBalance.GetBalance)
				r.Post("/withdraw", handlers.ForBalance.RequestWithdraw)
			})
			r.Get("/withdrawals", handlers.ForBalance.GetWithdrawals)
		})
	})

	return r
}
