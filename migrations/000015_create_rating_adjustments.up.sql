CREATE TABLE IF NOT EXISTS rating_adjustments (
    id BIGSERIAL PRIMARY KEY,
    actor_player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    target_player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    context_key TEXT NOT NULL,
    previous_rating INT NOT NULL,
    new_rating INT NOT NULL,
    previous_uncertainty INT NOT NULL,
    new_uncertainty INT NOT NULL,
    previous_matches_played INT NOT NULL,
    new_matches_played INT NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_rating_adjustments_actor_target_differ CHECK (actor_player_id <> target_player_id),
    CONSTRAINT ck_rating_adjustments_new_rating_nonnegative CHECK (new_rating >= 0),
    CONSTRAINT ck_rating_adjustments_new_uncertainty_nonnegative CHECK (new_uncertainty >= 0),
    CONSTRAINT ck_rating_adjustments_new_matches_nonnegative CHECK (new_matches_played >= 0)
);

CREATE INDEX IF NOT EXISTS ix_rating_adjustments_target_created
    ON rating_adjustments(target_player_id, created_at DESC);

