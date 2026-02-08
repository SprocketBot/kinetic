package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type migration struct {
	Version  int64
	Name     string
	Path     string
	Contents string
	Checksum string
}

type Migrator struct {
	db  *sql.DB
	dir string
}

func NewMigrator(db *sql.DB, migrationsDir string) *Migrator {
	return &Migrator{db: db, dir: migrationsDir}
}

func (m *Migrator) Up(ctx context.Context) (int, error) {
	if err := ensureMigrationsTable(ctx, m.db); err != nil {
		return 0, err
	}

	migrations, err := loadMigrationsFromDir(m.dir)
	if err != nil {
		return 0, err
	}

	applied := 0
	for _, migration := range migrations {
		exists, err := hasVersion(ctx, m.db, migration.Version)
		if err != nil {
			return applied, err
		}
		if exists {
			continue
		}

		if err := applyMigration(ctx, m.db, migration); err != nil {
			return applied, err
		}
		applied++
	}

	return applied, nil
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	const stmt = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    checksum TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`
	_, err := db.ExecContext(ctx, stmt)
	return err
}

func hasVersion(ctx context.Context, db *sql.DB, version int64) (bool, error) {
	const stmt = `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`
	var exists bool
	err := db.QueryRowContext(ctx, stmt, version).Scan(&exists)
	return exists, err
}

func applyMigration(ctx context.Context, db *sql.DB, migration migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, migration.Contents); err != nil {
		return fmt.Errorf("execute migration %d (%s): %w", migration.Version, migration.Name, err)
	}

	const stmt = `INSERT INTO schema_migrations(version, name, checksum) VALUES ($1, $2, $3)`
	if _, err = tx.ExecContext(ctx, stmt, migration.Version, migration.Name, migration.Checksum); err != nil {
		return fmt.Errorf("record migration %d (%s): %w", migration.Version, migration.Name, err)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func loadMigrationsFromDir(migrationsDir string) ([]migration, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	migrations := make([]migration, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		migration, err := parseMigrationFile(entry, migrationsDir)
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func parseMigrationFile(entry fs.DirEntry, migrationsDir string) (migration, error) {
	fileName := entry.Name()
	parts := strings.SplitN(fileName, "_", 2)
	if len(parts) != 2 {
		return migration{}, fmt.Errorf("invalid migration filename: %s", fileName)
	}

	version, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return migration{}, fmt.Errorf("invalid migration version for file %s: %w", fileName, err)
	}

	name := strings.TrimSuffix(parts[1], ".up.sql")
	fullPath := filepath.Join(migrationsDir, fileName)
	raw, err := os.ReadFile(fullPath)
	if err != nil {
		return migration{}, err
	}

	sum := sha256.Sum256(raw)
	checksum := hex.EncodeToString(sum[:])

	return migration{
		Version:  version,
		Name:     name,
		Path:     fullPath,
		Contents: string(raw),
		Checksum: checksum,
	}, nil
}
