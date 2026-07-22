DROP INDEX IF EXISTS ix_leagues_game_id;
ALTER TABLE leagues DROP COLUMN IF EXISTS game_id;
