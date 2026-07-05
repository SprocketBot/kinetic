ALTER TABLE result_submissions
    ADD COLUMN IF NOT EXISTS game_key TEXT NOT NULL DEFAULT 'rocket_league',
    ADD COLUMN IF NOT EXISTS submitted_by_subject TEXT NOT NULL DEFAULT 'legacy',
    ADD COLUMN IF NOT EXISTS submitted_by_display_name TEXT NOT NULL DEFAULT 'Legacy submission',
    ADD COLUMN IF NOT EXISTS payload_digest TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS provenance_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS home_ratified_by_subject TEXT NULL,
    ADD COLUMN IF NOT EXISTS home_ratified_by_display_name TEXT NULL,
    ADD COLUMN IF NOT EXISTS away_ratified_by_subject TEXT NULL,
    ADD COLUMN IF NOT EXISTS away_ratified_by_display_name TEXT NULL;

ALTER TABLE result_submissions
    DROP CONSTRAINT IF EXISTS ck_result_submissions_game_key_allowed;

ALTER TABLE result_submissions
    ADD CONSTRAINT ck_result_submissions_game_key_allowed
    CHECK (game_key IN ('rocket_league'));

CREATE INDEX IF NOT EXISTS ix_result_submissions_game_key_created_at
    ON result_submissions(game_key, created_at DESC);
