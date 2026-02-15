CREATE TABLE IF NOT EXISTS replay_evidence (
    id BIGSERIAL PRIMARY KEY,
    context_type TEXT NOT NULL,
    context_id BIGINT NOT NULL,
    submitted_by_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    replay_sha256 TEXT NOT NULL,
    content_size_bytes BIGINT NOT NULL,
    storage_ref TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'parsed',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_replay_evidence_context_type_allowed CHECK (context_type IN ('scrim', 'match')),
    CONSTRAINT ck_replay_evidence_state_allowed CHECK (state IN ('received', 'parsed', 'needs_review', 'finalized')),
    CONSTRAINT ck_replay_evidence_content_size_nonnegative CHECK (content_size_bytes >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_replay_evidence_sha256
    ON replay_evidence(replay_sha256);

CREATE INDEX IF NOT EXISTS ix_replay_evidence_created_at
    ON replay_evidence(created_at DESC);

CREATE TABLE IF NOT EXISTS replay_parse_runs (
    id BIGSERIAL PRIMARY KEY,
    replay_evidence_id BIGINT NOT NULL REFERENCES replay_evidence(id) ON DELETE CASCADE,
    parser_name TEXT NOT NULL,
    parser_version TEXT NOT NULL,
    parser_config_digest TEXT NOT NULL,
    status TEXT NOT NULL,
    output_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_replay_parse_runs_status_allowed CHECK (status IN ('parsed', 'failed', 'needs_review'))
);

CREATE INDEX IF NOT EXISTS ix_replay_parse_runs_evidence_created
    ON replay_parse_runs(replay_evidence_id, created_at DESC);

CREATE TABLE IF NOT EXISTS result_submission_replay_links (
    id BIGSERIAL PRIMARY KEY,
    result_submission_id BIGINT NOT NULL REFERENCES result_submissions(id) ON DELETE CASCADE,
    replay_evidence_id BIGINT NOT NULL REFERENCES replay_evidence(id) ON DELETE CASCADE,
    linked_by_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_result_submission_replay_link UNIQUE (result_submission_id, replay_evidence_id)
);

CREATE INDEX IF NOT EXISTS ix_result_submission_replay_links_submission
    ON result_submission_replay_links(result_submission_id, created_at DESC);
