CREATE TABLE IF NOT EXISTS player_ratings (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    context_key TEXT NOT NULL,
    rating INT NOT NULL DEFAULT 1000,
    uncertainty INT NOT NULL DEFAULT 350,
    matches_played INT NOT NULL DEFAULT 0,
    last_competed_at TIMESTAMPTZ NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_player_ratings_active_identity
    ON player_ratings(player_id, context_key)
    WHERE is_active;

ALTER TABLE queue_entries
    ADD COLUMN IF NOT EXISTS expansion_stage INT NOT NULL DEFAULT 1;

ALTER TABLE queue_entries
    ADD COLUMN IF NOT EXISTS stage_advanced_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_queue_entries_expansion_stage_positive'
    ) THEN
        ALTER TABLE queue_entries
            ADD CONSTRAINT ck_queue_entries_expansion_stage_positive
            CHECK (expansion_stage >= 1);
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS scrims (
    id BIGSERIAL PRIMARY KEY,
    queue_id BIGINT NOT NULL REFERENCES queues(id) ON DELETE CASCADE,
    home_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    away_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    state TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ NULL,
    ended_at TIMESTAMPTZ NULL,
    CONSTRAINT ck_scrims_teams_differ CHECK (home_team_id <> away_team_id),
    CONSTRAINT ck_scrims_state_allowed CHECK (state IN ('created', 'in_progress', 'closed', 'voided'))
);

CREATE TABLE IF NOT EXISTS matchmaking_decisions (
    id BIGSERIAL PRIMARY KEY,
    scrim_id BIGINT NOT NULL REFERENCES scrims(id) ON DELETE CASCADE,
    queue_id BIGINT NOT NULL REFERENCES queues(id) ON DELETE CASCADE,
    queue_wait_seconds INT NOT NULL,
    expansion_stage INT NOT NULL,
    rating_spread INT NOT NULL DEFAULT 0,
    cross_group BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
