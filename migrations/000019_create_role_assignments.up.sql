CREATE TABLE IF NOT EXISTS role_assignments (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    franchise_id BIGINT NULL REFERENCES franchises(id) ON DELETE CASCADE,
    club_id BIGINT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    team_id BIGINT NULL REFERENCES teams(id) ON DELETE CASCADE,
    assigned_by_actor_player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE RESTRICT,
    assign_reason TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_by_actor_player_id BIGINT NULL REFERENCES players(id) ON DELETE RESTRICT,
    revoke_reason TEXT NULL,
    revoked_at TIMESTAMPTZ NULL,
    CONSTRAINT ck_role_assignments_role CHECK (role IN ('fm', 'gm', 'agm', 'captain')),
    CONSTRAINT ck_role_assignments_scope_consistent CHECK (
        (role = 'fm' AND franchise_id IS NOT NULL AND club_id IS NULL AND team_id IS NULL)
        OR
        (role IN ('gm', 'agm') AND franchise_id IS NULL AND club_id IS NOT NULL AND team_id IS NULL)
        OR
        (role = 'captain' AND franchise_id IS NULL AND club_id IS NULL AND team_id IS NOT NULL)
    ),
    CONSTRAINT ck_role_assignments_revoke_state_consistent CHECK (
        (is_active = TRUE AND revoked_by_actor_player_id IS NULL AND revoke_reason IS NULL AND revoked_at IS NULL)
        OR
        (is_active = FALSE AND revoked_by_actor_player_id IS NOT NULL AND revoke_reason IS NOT NULL AND revoked_at IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_role_assignments_active_identity
    ON role_assignments(
        player_id,
        role,
        COALESCE(franchise_id, 0),
        COALESCE(club_id, 0),
        COALESCE(team_id, 0)
    )
    WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS ix_role_assignments_role_scope
    ON role_assignments(role, franchise_id, club_id, team_id, is_active);

