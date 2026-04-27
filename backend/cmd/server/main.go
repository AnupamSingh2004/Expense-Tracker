package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fenmo/expense-tracker/internal/config"
	"github.com/fenmo/expense-tracker/internal/db"
	"github.com/fenmo/expense-tracker/internal/handler"
	"github.com/fenmo/expense-tracker/internal/middleware"
	"github.com/fenmo/expense-tracker/internal/repository"
	"github.com/fenmo/expense-tracker/internal/service"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	if err := runMigrations(cfg, logger); err != nil {
		logger.Fatal("migrations failed", zap.Error(err))
	}

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DSN())
	if err != nil {
		logger.Fatal("db connect failed", zap.Error(err))
	}
	defer pool.Close()

	expRepo := repository.NewExpenseRepository(pool)
	idempRepo := repository.NewIdempotencyRepository(pool)
	expSvc := service.NewExpenseService(expRepo, idempRepo, logger)
	expHandler := handler.NewExpenseHandler(expSvc, logger)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Timeout(10 * time.Second))
	r.Use(chimw.Recoverer)

	r.Get("/health", handler.Health)
	r.Route("/expenses", func(r chi.Router) {
		r.Post("/", expHandler.Create)
		r.Get("/", expHandler.List)
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.ServerPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
	}
	logger.Info("server stopped")
}

func runMigrations(cfg *config.Config, logger *zap.Logger) error {
	m, err := migrate.New("file://"+cfg.MigrationsPath, cfg.PostgresDSN())
	if err != nil {
		return fmt.Errorf("migration init: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up: %w", err)
	}
	logger.Info("migrations applied")
	return nil
}
