ALTER TABLE rating_adjustments
    DROP CONSTRAINT IF EXISTS ck_rating_adjustments_actor_target_differ;

ALTER TABLE rating_adjustments
    ADD CONSTRAINT ck_rating_adjustments_actor_target_differ
        CHECK (actor_player_id <> target_player_id);

ALTER TABLE rating_adjustments
    ALTER COLUMN actor_player_id SET NOT NULL;
