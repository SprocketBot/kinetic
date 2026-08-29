package http

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var integrationSchemaSequence atomic.Uint64

func isolatedTestDatabaseURL(t *testing.T) string {
	t.Helper()
	baseURL := os.Getenv("TEST_DATABASE_URL")
	if baseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	schema := fmt.Sprintf("test_%d_%d", time.Now().UnixNano(), integrationSchemaSequence.Add(1))
	admin, err := sql.Open("pgx", baseURL)
	if err != nil {
		t.Fatalf("open test database for schema setup: %v", err)
	}
	if _, err := admin.Exec(`CREATE SCHEMA ` + schema); err != nil {
		_ = admin.Close()
		t.Fatalf("create isolated test schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(`DROP SCHEMA IF EXISTS ` + schema + ` CASCADE`)
		_ = admin.Close()
	})

	parsed, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
