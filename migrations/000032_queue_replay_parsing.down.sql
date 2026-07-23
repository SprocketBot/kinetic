DROP INDEX IF EXISTS ix_replay_parse_runs_queued;
ALTER TABLE replay_parse_runs
    DROP CONSTRAINT ck_replay_parse_runs_status_allowed,
    DROP COLUMN finished_at,
    DROP COLUMN started_at,
    DROP COLUMN failure_reason,
    ADD CONSTRAINT ck_replay_parse_runs_status_allowed CHECK (status IN ('parsed', 'failed', 'needs_review'));
ALTER TABLE replay_evidence DROP COLUMN replay_body;
