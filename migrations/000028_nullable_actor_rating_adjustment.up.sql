-- Allow system-automated rating adjustments where actor_player_id is NULL.
ALTER TABLE rating_adjustments
    ALTER COLUMN actor_player_id DROP NOT NULL;

ALTER TABLE rating_adjustments
    DROP CONSTRAINT IF EXISTS ck_rating_adjustments_actor_target_differ;

ALTER TABLE rating_adjustments
    ADD CONSTRAINT ck_rating_adjustments_actor_target_differ
        CHECK (actor_player_id IS NULL OR actor_player_id <> target_player_id);
