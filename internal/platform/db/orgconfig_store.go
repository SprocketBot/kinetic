package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/kineticbot/kinetic-v3/internal/domain/orgconfig"
)

type OrgConfigStore struct {
	db *sql.DB
}

func NewOrgConfigStore(db *sql.DB) *OrgConfigStore {
	return &OrgConfigStore{db: db}
}

func (s *OrgConfigStore) GetConfig(ctx context.Context, leagueID int64, key string) (string, error) {
	const stmt = `
SELECT value
FROM organization_configs
WHERE league_id = $1 AND key = $2;`

	var value string
	err := s.db.QueryRowContext(ctx, stmt, leagueID, key).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return orgconfig.Default(key), nil
		}
		return "", err
	}
	return value, nil
}

func (s *OrgConfigStore) ListConfigs(ctx context.Context, leagueID int64) (map[string]string, error) {
	const stmt = `
SELECT key, value
FROM organization_configs
WHERE league_id = $1;`

	rows, err := s.db.QueryContext(ctx, stmt, leagueID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		result[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, k := range orgconfig.AllKnownKeys() {
		if _, ok := result[k]; !ok {
			result[k] = orgconfig.Default(k)
		}
	}

	return result, nil
}

func (s *OrgConfigStore) SetConfig(ctx context.Context, leagueID int64, key, value string, updatedBy *int64) error {
	const stmt = `
INSERT INTO organization_configs (league_id, key, value, updated_by, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (league_id, key) DO UPDATE
    SET value      = EXCLUDED.value,
        updated_by = EXCLUDED.updated_by,
        updated_at = NOW();`

	var nullUpdatedBy sql.NullInt64
	if updatedBy != nil {
		nullUpdatedBy = sql.NullInt64{Int64: *updatedBy, Valid: true}
	}

	_, err := s.db.ExecContext(ctx, stmt, leagueID, key, value, nullUpdatedBy)
	if err != nil {
		return err
	}
	return nil
}
