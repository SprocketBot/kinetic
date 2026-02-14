ALTER TABLE players
    DROP CONSTRAINT IF EXISTS uq_player_team_display_name;

ALTER TABLE players
    DROP COLUMN IF EXISTS team_id;

DROP INDEX IF EXISTS uq_roster_memberships_active_pair;

CREATE UNIQUE INDEX IF NOT EXISTS uq_roster_memberships_active_player
    ON roster_memberships(player_id)
    WHERE is_active;
