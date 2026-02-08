package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMigratorUpIsRepeatable(t *testing.T) {
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	dbConn, err := Open(testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	defer dbConn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := Ping(ctx, dbConn); err != nil {
		t.Fatalf("failed to ping test DB: %v", err)
	}

	tableName := fmt.Sprintf("integration_migration_check_%d", time.Now().UnixNano())
	dir := t.TempDir()
	writeIntegrationMigration(t, dir, "000001_create_table.up.sql", fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (id BIGINT PRIMARY KEY);`, tableName))

	migrator := NewMigrator(dbConn, dir)

	firstRunApplied, err := migrator.Up(ctx)
	if err != nil {
		t.Fatalf("first migration run failed: %v", err)
	}
	if firstRunApplied != 1 {
		t.Fatalf("expected first run to apply 1 migration, got %d", firstRunApplied)
	}

	secondRunApplied, err := migrator.Up(ctx)
	if err != nil {
		t.Fatalf("second migration run failed: %v", err)
	}
	if secondRunApplied != 0 {
		t.Fatalf("expected second run to apply 0 migrations, got %d", secondRunApplied)
	}

	if !tableExists(ctx, t, dbConn, tableName) {
		t.Fatalf("expected migrated table %s to exist", tableName)
	}
}

func writeIntegrationMigration(t *testing.T, dir, name, contents string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write integration migration %s: %v", name, err)
	}
}

func tableExists(ctx context.Context, t *testing.T, conn *sql.DB, tableName string) bool {
	t.Helper()

	var exists bool
	err := conn.QueryRowContext(
		ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1)`,
		tableName,
	).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to query table existence: %v", err)
	}
	return exists
}
