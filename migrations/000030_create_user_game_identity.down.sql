DROP INDEX IF EXISTS ix_platform_account_links_player_id;
ALTER TABLE platform_account_links DROP COLUMN IF EXISTS player_id;
DROP TABLE IF EXISTS user_players;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS games;
