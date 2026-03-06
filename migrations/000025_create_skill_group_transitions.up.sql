CREATE TABLE IF NOT EXISTS skill_group_transitions (
    id                    BIGSERIAL PRIMARY KEY,
    player_id             BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    from_skill_group_id   BIGINT REFERENCES skill_groups(id) ON DELETE SET NULL,
    to_skill_group_id     BIGINT NOT NULL REFERENCES skill_groups(id) ON DELETE RESTRICT,
    rating_at_transition  INT NOT NULL,
    direction             TEXT NOT NULL,
    effective_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_sgt_direction CHECK (direction IN ('promotion', 'demotion'))
);

CREATE INDEX IF NOT EXISTS ix_sgt_player_effective
    ON skill_group_transitions(player_id, effective_at DESC);
