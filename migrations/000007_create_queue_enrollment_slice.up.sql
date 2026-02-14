CREATE TABLE IF NOT EXISTS queues (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_queue_name UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS queue_entries (
    id BIGSERIAL PRIMARY KEY,
    queue_id BIGINT NOT NULL REFERENCES queues(id) ON DELETE CASCADE,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_queue_entries_active_queue_team
    ON queue_entries(queue_id, team_id)
    WHERE is_active;

CREATE INDEX IF NOT EXISTS ix_queue_entries_active_ordering
    ON queue_entries(queue_id, is_active, created_at, id);
