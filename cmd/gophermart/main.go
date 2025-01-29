package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Melikhov-p/go-loyalty-system/internal/config"
	"github.com/Melikhov-p/go-loyalty-system/internal/logger"
	"github.com/Melikhov-p/go-loyalty-system/internal/router"
	"github.com/Melikhov-p/go-loyalty-system/internal/workers"
	"github.com/Melikhov-p/go-loyalty-system/pkg"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	maxWorkersCount = 10

	timeoutServerShutdown = time.Second * 5
	timeoutShutdown       = time.Second * 10
)

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}

func Run() (err error) {
	rootCtx, cancelCtx := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelCtx()

	eg, ctx := errgroup.WithContext(rootCtx)
	// нештатное завершение программы по таймауту
	// происходит, если после завершения контекста
	// приложение не смогло завершиться за отведенный промежуток времени
	context.AfterFunc(ctx, func() {
		ctx, cancelCtx := context.WithTimeout(context.Background(), timeoutShutdown)
		defer cancelCtx()

		<-ctx.Done()
		log.Fatal("failed to gracefully shutdown the service")
	})

	cfg := config.BuildConfig()

	lgr, err := logger.BuildLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("error building logger: %w", err)
	}
	lgr.Info("config and logger ready", zap.Any("CONFIG", cfg))

	db, err := pkg.ConnectDB(cfg)
	if err != nil {
		return fmt.Errorf("error connecting database: %w", err)
	}
	lgr.Debug("database connected")
	// отслеживаем успешное закрытие соединения с БД
	eg.Go(func() error {
		defer log.Print("closed DB")

		<-ctx.Done()

		_ = db.Close()
		return nil
	})

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

	eg.Go(func() error {
		if err = httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server can not listen anymore: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		defer lgr.Debug("server has been shutdown")
		<-ctx.Done()

		shutdownTimeoutCtx, cancelShutdownTimeoutCtx := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancelShutdownTimeoutCtx()
		if err = httpServer.Shutdown(shutdownTimeoutCtx); err != nil {
			log.Printf("an error occurred during server shutdown: %v", err)
		}
		return nil
	})

	workDispatcher := workers.NewDispatcher(lgr, maxWorkersCount, cfg, db, cfg.Dispatcher.PingInterval)

	eg.Go(func() error {
		workDispatcher.Run()
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()

		workDispatcher.Stop()
		return nil
	})

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("errgroup error: %w", err)
	}

	return nil
}
