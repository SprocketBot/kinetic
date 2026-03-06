CREATE TABLE IF NOT EXISTS api_tokens (
    id            BIGSERIAL PRIMARY KEY,
    name          TEXT NOT NULL,
    token_hash    TEXT NOT NULL UNIQUE,
    subject       TEXT NOT NULL,
    scopes        TEXT[] NOT NULL DEFAULT '{}',
    created_by    BIGINT REFERENCES players(id) ON DELETE SET NULL,
    expires_at    TIMESTAMPTZ,
    revoked_at    TIMESTAMPTZ,
    last_used_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_api_tokens_subject
    ON api_tokens(subject)
    WHERE revoked_at IS NULL;
