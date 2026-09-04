package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vtech/our-sell/internal/app"
	"github.com/vtech/our-sell/internal/config"
	"github.com/vtech/our-sell/internal/database"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()
	postgres, err := database.NewPostgres(startupCtx, cfg)
	if err != nil {
		logger.Error("database startup failed", "error", err)
		os.Exit(1)
	}
	defer postgres.Close()
	redisClient, err := database.NewRedis(startupCtx, cfg.RedisURL)
	if err != nil {
		logger.Error("Redis startup failed", "error", err)
		os.Exit(1)
	}
	defer func() { _ = redisClient.Close() }()

	application, err := app.New(app.Dependencies{Config: cfg, Logger: logger, Postgres: postgres, Redis: redisClient})
	if err != nil {
		logger.Error("application startup failed", "error", err)
		os.Exit(1)
	}

	serverErr := make(chan error, 1)
	go func() { serverErr <- application.Fiber.Listen(":" + cfg.AppPort) }()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErr:
		logger.Error("HTTP server stopped unexpectedly", "error", err)
		os.Exit(1)
	case <-shutdownSignal.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := application.Fiber.ShutdownWithContext(shutdownCtx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
		logger.Info("server stopped")
	}
}
