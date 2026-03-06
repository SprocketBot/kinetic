CREATE TABLE IF NOT EXISTS player_notifications (
    id           BIGSERIAL PRIMARY KEY,
    player_id    BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    category     TEXT NOT NULL,
    context_type TEXT,
    context_id   BIGINT,
    message      TEXT NOT NULL,
    read_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_player_notifications_category CHECK (
        category IN ('scrim_popped', 'queue_ban', 'ratification_needed', 'skill_group_change', 'roster_change')
    )
);

CREATE INDEX IF NOT EXISTS ix_player_notifications_player_unread
    ON player_notifications(player_id, created_at DESC)
    WHERE read_at IS NULL;
