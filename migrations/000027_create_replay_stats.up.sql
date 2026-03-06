CREATE TABLE IF NOT EXISTS rounds (
    id              BIGSERIAL PRIMARY KEY,
    submission_id   BIGINT NOT NULL REFERENCES result_submissions(id) ON DELETE CASCADE,
    round_number    INT NOT NULL,
    duration_seconds INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_rounds_submission_number UNIQUE (submission_id, round_number)
);

CREATE INDEX IF NOT EXISTS ix_rounds_submission_id
    ON rounds(submission_id);

CREATE TABLE IF NOT EXISTS player_stat_lines (
    id              BIGSERIAL PRIMARY KEY,
    round_id        BIGINT NOT NULL REFERENCES rounds(id) ON DELETE CASCADE,
    player_id       BIGINT REFERENCES players(id) ON DELETE SET NULL,
    replay_identity TEXT NOT NULL,
    goals           INT NOT NULL DEFAULT 0,
    assists         INT NOT NULL DEFAULT 0,
    saves           INT NOT NULL DEFAULT 0,
    shots           INT NOT NULL DEFAULT 0,
    score           INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_player_stat_lines_round_id
    ON player_stat_lines(round_id);

CREATE INDEX IF NOT EXISTS ix_player_stat_lines_player_id
    ON player_stat_lines(player_id)
    WHERE player_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS team_stat_lines (
    id       BIGSERIAL PRIMARY KEY,
    round_id BIGINT NOT NULL REFERENCES rounds(id) ON DELETE CASCADE,
    team_id  BIGINT REFERENCES teams(id) ON DELETE SET NULL,
    goals    INT NOT NULL DEFAULT 0,
    shots    INT NOT NULL DEFAULT 0,
    saves    INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_team_stat_lines_round_id
    ON team_stat_lines(round_id);
