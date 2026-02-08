package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sprocketbot/sprocket-v3/internal/domain/authz"
	"github.com/sprocketbot/sprocket-v3/internal/platform/auth"
	"github.com/sprocketbot/sprocket-v3/internal/platform/config"
	"github.com/sprocketbot/sprocket-v3/internal/platform/db"
	httpserver "github.com/sprocketbot/sprocket-v3/internal/platform/http"
)

func main() {
	cfg := config.Load()
	logger := newLogger(cfg.LogLevel)
	ctx := context.Background()
	var conn *sql.DB

	if cfg.RequireDatabase || cfg.RunMigrationsOnStart {
		var err error
		conn, err = db.Open(cfg.DatabaseURL)
		if err != nil {
			logger.Error("failed to open database", "error", err)
			os.Exit(1)
		}
		defer conn.Close()

		if err := db.Ping(ctx, conn); err != nil {
			logger.Error("failed to ping database", "error", err)
			os.Exit(1)
		}

		logger.Info("database connectivity validated")

		if cfg.RunMigrationsOnStart {
			migrator := db.NewMigrator(conn, cfg.MigrationsDir)
			applied, err := migrator.Up(ctx)
			if err != nil {
				logger.Error("failed to run migrations on startup", "error", err)
				os.Exit(1)
			}
			logger.Info("startup migrations complete", "applied", applied)
		}
	}

	tokenValidator := auth.NewLocalTokenValidator("local")
	var evaluator authz.Evaluator
	if conn != nil {
		dbEvaluator, err := authz.NewDatabaseBackedEvaluator(ctx, conn, authz.DefaultPermissions())
		if err != nil {
			logger.Error("failed to load authz policies from database", "error", err)
			os.Exit(1)
		}
		logger.Info("loaded authz policies from database")
		evaluator = dbEvaluator
	} else {
		staticEvaluator := authz.NewStaticEvaluator(authz.DefaultPermissions())
		evaluator = staticEvaluator
		logger.Warn("using static in-memory authz policy set; database not required in current mode")
	}

	srv := httpserver.New(cfg, logger, httpserver.Dependencies{
		TokenValidator: tokenValidator,
		Evaluator:      evaluator,
	})

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}

func newLogger(level string) *slog.Logger {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slogLevel})
	return slog.New(handler)
}
