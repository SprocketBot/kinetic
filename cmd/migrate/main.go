package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/kineticbot/kinetic-v3/internal/platform/config"
	"github.com/kineticbot/kinetic-v3/internal/platform/db"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	conn, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	if err := db.Ping(ctx, conn); err != nil {
		logger.Error("failed to ping database", "error", err)
		os.Exit(1)
	}

	migrator := db.NewMigrator(conn, cfg.MigrationsDir)
	applied, err := migrator.Up(ctx)
	if err != nil {
		logger.Error("migration run failed", "error", err)
		os.Exit(1)
	}

	logger.Info("migration run complete", "applied", applied)
}
