package authz

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kineticbot/kinetic-v3/internal/platform/db"
)

func TestLoadPermissionsFromDatabase(t *testing.T) {
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	conn, err := db.Open(testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := db.Ping(ctx, conn); err != nil {
		t.Fatalf("failed to ping test DB: %v", err)
	}

	migrator := db.NewMigrator(conn, "../../../migrations")
	if _, err := migrator.Up(ctx); err != nil {
		t.Fatalf("failed to migrate test DB: %v", err)
	}

	permissions, err := LoadPermissions(ctx, conn)
	if err != nil {
		t.Fatalf("failed to load permissions: %v", err)
	}

	if len(permissions) == 0 {
		t.Fatal("expected at least one permission from migrated seed data")
	}
}
