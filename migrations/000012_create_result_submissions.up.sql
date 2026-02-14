CREATE TABLE IF NOT EXISTS result_submissions (
    id BIGSERIAL PRIMARY KEY,
    context_type TEXT NOT NULL,
    context_id BIGINT NOT NULL,
    submitted_by_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    home_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    away_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    winning_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    losing_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    state TEXT NOT NULL DEFAULT 'pending',
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    home_ratified_at TIMESTAMPTZ NULL,
    away_ratified_at TIMESTAMPTZ NULL,
    rejected_by_team_id BIGINT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    rejection_reason TEXT NULL,
    rejected_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_result_submissions_context_type_allowed CHECK (context_type IN ('scrim', 'match')),
    CONSTRAINT ck_result_submissions_state_allowed CHECK (state IN ('pending', 'ratified', 'rejected')),
    CONSTRAINT ck_result_submissions_teams_differ CHECK (home_team_id <> away_team_id),
    CONSTRAINT ck_result_submissions_winner_loser_differ CHECK (winning_team_id <> losing_team_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_result_submissions_pending_context
    ON result_submissions(context_type, context_id)
    WHERE state = 'pending';

CREATE INDEX IF NOT EXISTS ix_result_submissions_created_at
    ON result_submissions(created_at DESC);
