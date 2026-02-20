CREATE TABLE IF NOT EXISTS platform_account_links (
    id BIGSERIAL PRIMARY KEY,
    subject TEXT NOT NULL,
    provider TEXT NOT NULL,
    provider_account_id TEXT NOT NULL,
    provider_account_name TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    unlinked_at TIMESTAMPTZ NULL,
    CONSTRAINT ck_platform_account_links_provider CHECK (provider IN ('steam', 'xbox', 'psn', 'epic'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_platform_account_links_active_provider_account
    ON platform_account_links(provider, provider_account_id)
    WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS ix_platform_account_links_subject
    ON platform_account_links(subject, is_active, linked_at DESC);
