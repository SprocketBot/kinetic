package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMigrationsFromDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeTestFile(t, dir, "000002_second.up.sql", "SELECT 1;")
	writeTestFile(t, dir, "000001_first.up.sql", "SELECT 1;")
	writeTestFile(t, dir, "README.md", "ignore me")

	migrations, err := loadMigrationsFromDir(dir)
	if err != nil {
		t.Fatalf("loadMigrationsFromDir failed: %v", err)
	}

	if len(migrations) != 2 {
		t.Fatalf("expected 2 migrations, got %d", len(migrations))
	}

	if migrations[0].Version != 1 || migrations[1].Version != 2 {
		t.Fatalf("migrations not sorted by version: %+v", migrations)
	}
}

func TestLoadMigrationsFromDirRejectsInvalidFilename(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeTestFile(t, dir, "invalid_name.up.sql", "SELECT 1;")

	_, err := loadMigrationsFromDir(dir)
	if err == nil {
		t.Fatal("expected error for invalid migration filename")
	}
}

func writeTestFile(t *testing.T, dir, name, contents string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("failed to write test file %s: %v", name, err)
	}
}
