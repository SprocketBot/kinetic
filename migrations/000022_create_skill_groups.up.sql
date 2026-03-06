CREATE TABLE IF NOT EXISTS skill_groups (
    id                   BIGSERIAL PRIMARY KEY,
    league_id            BIGINT NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
    name                 TEXT NOT NULL,
    display_order        INT NOT NULL,
    rating_floor         INT NOT NULL,
    rating_ceiling       INT NOT NULL,
    promotion_threshold  INT,
    demotion_threshold   INT,
    is_active            BOOLEAN NOT NULL DEFAULT TRUE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_skill_groups_league_name UNIQUE (league_id, name),
    CONSTRAINT ck_skill_groups_floor_lte_ceiling CHECK (rating_floor <= rating_ceiling),
    CONSTRAINT ck_skill_groups_promotion_above_ceiling CHECK (promotion_threshold IS NULL OR promotion_threshold > rating_ceiling),
    CONSTRAINT ck_skill_groups_demotion_below_floor CHECK (demotion_threshold IS NULL OR demotion_threshold < rating_floor)
);

CREATE INDEX IF NOT EXISTS ix_skill_groups_league_order
    ON skill_groups(league_id, display_order);
