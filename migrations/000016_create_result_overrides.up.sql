CREATE TABLE IF NOT EXISTS result_overrides (
    id BIGSERIAL PRIMARY KEY,
    submission_id BIGINT NOT NULL REFERENCES result_submissions(id) ON DELETE CASCADE,
    actor TEXT NOT NULL,
    reason TEXT NOT NULL,
    previous_winning_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    previous_losing_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    new_winning_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    new_losing_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    previous_state TEXT NOT NULL,
    new_state TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_result_overrides_previous_teams_differ CHECK (previous_winning_team_id <> previous_losing_team_id),
    CONSTRAINT ck_result_overrides_new_teams_differ CHECK (new_winning_team_id <> new_losing_team_id)
);

CREATE INDEX IF NOT EXISTS ix_result_overrides_submission_created
    ON result_overrides(submission_id, created_at DESC);
