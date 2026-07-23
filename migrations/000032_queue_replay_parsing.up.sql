ALTER TABLE replay_evidence ADD COLUMN replay_body TEXT NOT NULL DEFAULT '';

ALTER TABLE replay_parse_runs
    DROP CONSTRAINT ck_replay_parse_runs_status_allowed,
    ADD COLUMN failure_reason TEXT,
    ADD COLUMN started_at TIMESTAMPTZ,
    ADD COLUMN finished_at TIMESTAMPTZ,
    ADD CONSTRAINT ck_replay_parse_runs_status_allowed
        CHECK (status IN ('queued', 'running', 'parsed', 'failed', 'needs_review'));

CREATE INDEX ix_replay_parse_runs_queued ON replay_parse_runs(created_at ASC) WHERE status = 'queued';
