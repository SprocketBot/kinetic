ALTER TABLE scrims
    DROP CONSTRAINT IF EXISTS ck_scrims_state_allowed;

ALTER TABLE scrims
    ADD CONSTRAINT ck_scrims_state_allowed
    CHECK (state IN ('created', 'popped', 'in_progress', 'closed', 'voided', 'cancelled'));

ALTER TABLE scrims
    ADD COLUMN IF NOT EXISTS lobby_name           TEXT,
    ADD COLUMN IF NOT EXISTS lobby_password       TEXT,
    ADD COLUMN IF NOT EXISTS popped_at            TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS home_checked_in_at   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS away_checked_in_at   TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS ix_scrims_active_state
    ON scrims(state)
    WHERE state NOT IN ('closed', 'voided', 'cancelled');
