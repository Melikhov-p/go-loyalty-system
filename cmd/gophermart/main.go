package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/logger"
	"github.com/Melikhov-p/go-loyalty-system/internal/router"
	"github.com/Melikhov-p/go-loyalty-system/internal/workers"
	"github.com/Melikhov-p/go-loyalty-system/pkg"
	"go.uber.org/zap"
)

func main() {
	cfg := config.BuildConfig()

	lgr, err := logger.BuildLogger(cfg.LogLevel)
	if err != nil {
		panic("error building logger: " + err.Error())
	}
	lgr.Debug("config and logger ready")

	db, err := pkg.ConnectDB(cfg)
	if err != nil {
		panic("error connecting database: " + err.Error())
	}
	lgr.Debug("database connected")

	r := router.CreateRouter(cfg, lgr, db)
	lgr.Debug("router ready")

	lgr.Debug("starting server",
		zap.String("RunAddr", cfg.RunAddr),
		zap.String("DatabaseURI", cfg.DB.DatabaseURI),
		zap.String("AccrualAddr", cfg.AccrualAddr),
		zap.String("LogLevel", cfg.LogLevel))

	httpServer := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: r,
	}

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err = httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			lgr.Fatal("server can not listen anymore", zap.Error(err))
		}
	}()

	updateOrdersWatcher := workers.NewOrderWatcher(db, lgr, cfg, cfg.Worker.PingInterval)

	go func() {
		updateOrdersWatcher.Work()
	}()

	<-stopCh
	lgr.Info("Stopping work worker")

	updateOrdersWatcher.Stop()
	lgr.Info("Worker stopped")

	lgr.Info("Shutting down server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err = httpServer.Shutdown(ctx); err != nil {
		lgr.Fatal("server forced shutdown", zap.Error(err))
	}

	lgr.Info("Server gracefully shutdown")
}
