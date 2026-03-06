DROP INDEX IF EXISTS ix_scrims_active_state;

ALTER TABLE scrims
    DROP COLUMN IF EXISTS away_checked_in_at,
    DROP COLUMN IF EXISTS home_checked_in_at,
    DROP COLUMN IF EXISTS popped_at,
    DROP COLUMN IF EXISTS lobby_password,
    DROP COLUMN IF EXISTS lobby_name;

ALTER TABLE scrims
    DROP CONSTRAINT IF EXISTS ck_scrims_state_allowed;

ALTER TABLE scrims
    ADD CONSTRAINT ck_scrims_state_allowed
    CHECK (state IN ('created', 'in_progress', 'closed', 'voided'));
