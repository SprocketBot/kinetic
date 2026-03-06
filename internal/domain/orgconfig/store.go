package orgconfig

import "context"

type Store interface {
	// GetConfig returns the value for key, or Default(key) if not set in DB.
	GetConfig(ctx context.Context, leagueID int64, key string) (string, error)
	// ListConfigs returns all config values for the league, merged with defaults for unset known keys.
	ListConfigs(ctx context.Context, leagueID int64) (map[string]string, error)
	// SetConfig upserts a config value.
	SetConfig(ctx context.Context, leagueID int64, key, value string, updatedBy *int64) error
}
