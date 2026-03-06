CREATE TABLE IF NOT EXISTS organization_configs (
    id          BIGSERIAL PRIMARY KEY,
    league_id   BIGINT NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    key         TEXT NOT NULL,
    value       TEXT NOT NULL,
    updated_by  BIGINT REFERENCES players(id) ON DELETE SET NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_organization_configs_league_key UNIQUE (league_id, key)
);

CREATE INDEX IF NOT EXISTS ix_organization_configs_league_id
    ON organization_configs(league_id);
