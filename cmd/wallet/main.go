package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	"go.uber.org/zap"

	"github.com/Te8va/wallet/internal/config"
	"github.com/Te8va/wallet/internal/handler"
	"github.com/Te8va/wallet/internal/middleware"
	"github.com/Te8va/wallet/internal/repository"
	"github.com/Te8va/wallet/internal/service"
)

func main() {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	defer func() {
		if err := logger.Sync(); err != nil {
			sugar.Errorw("Failed to sync logger", "error", err)
		}
	}()

	cfg := config.Config{}
	if err := env.Parse(&cfg); err != nil {
		logger.Fatal("Failed to parse env: %v", zap.Error(err))
	}

	m, err := migrate.New("file://migrations", cfg.PostgresConn)
	if err != nil {
		logger.Fatal("Failed to create migrate", zap.Error(err))
	}

	err = repository.ApplyMigrations(m)
	if err != nil {
		logger.Fatal("Failed to apply migrations", zap.Error(err))
	}

	logger.Info("Migrations applied successfully")

	ctx := context.Background()
	pool, err := repository.GetPgxPool(ctx, cfg.PostgresConn)
	if err != nil {
		logger.Fatal("Failed to create postgres connection pool", zap.Error(err))
	}

	defer pool.Close()

	logger.Info("Postgres connection pool created")

	var wg sync.WaitGroup

	walletRepo, err := repository.NewWalletRepository(pool)
	if err != nil {
		sugar.Fatalf("Failed to create wallet repository: %v", err)
	}
	walletService := service.NewWalletService(walletRepo)
	walletHandler := handler.NewWalletHandler(walletService)

	_, cancelDeleteCtx := context.WithCancel(context.Background())

	r := chi.NewRouter()
	r.Use(middleware.WithLogging)

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/wallet", walletHandler.WalletOperationHandler)

		r.Get("/wallets/{walletId}", walletHandler.GetBalanceHandler)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.ServiceHost, cfg.ServicePort),
		Handler: r,
	}

	go func() {
		logger.Info("Server started, listening on port", zap.Int("port", cfg.ServicePort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("ListenAndServe failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server shutdown failed:", zap.Error(err))
	}

	waitGroupChan := make(chan struct{})
	go func() {
		wg.Wait()
		waitGroupChan <- struct{}{}
	}()

	select {
	case <-waitGroupChan:
		logger.Info("All delete goroutines successfully finished")
	case <-time.After(time.Second * 3):
		cancelDeleteCtx()
		logger.Info("Some of delete goroutines have not completed their job due to shutdown timeout")
	}

	logger.Info("Server was shut down")
}
