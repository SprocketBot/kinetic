ALTER TABLE leagues ADD COLUMN IF NOT EXISTS game_id BIGINT NULL REFERENCES games(id) ON DELETE RESTRICT;

UPDATE leagues
SET game_id = (SELECT id FROM games WHERE slug = 'rocket-league')
WHERE game_id IS NULL;

ALTER TABLE leagues ALTER COLUMN game_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS ix_leagues_game_id ON leagues(game_id);
