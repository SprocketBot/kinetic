CREATE TABLE IF NOT EXISTS queue_bans (
    id BIGSERIAL PRIMARY KEY,
    queue_id BIGINT NOT NULL REFERENCES queues(id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    banned_by_actor TEXT NOT NULL,
    ban_reason TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    banned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    unbanned_by_actor TEXT NULL,
    unban_reason TEXT NULL,
    unbanned_at TIMESTAMPTZ NULL,
    CONSTRAINT ck_queue_bans_unban_state_consistent CHECK (
        (is_active = TRUE AND unbanned_by_actor IS NULL AND unban_reason IS NULL AND unbanned_at IS NULL)
        OR
        (is_active = FALSE AND unbanned_by_actor IS NOT NULL AND unban_reason IS NOT NULL AND unbanned_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_queue_bans_active_identity
    ON queue_bans(queue_id, player_id)
    WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS ix_queue_bans_player_queue
    ON queue_bans(player_id, queue_id);
