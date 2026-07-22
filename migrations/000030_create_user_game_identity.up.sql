CREATE TABLE IF NOT EXISTS games (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO games(name, slug)
VALUES ('Rocket League', 'rocket-league')
ON CONFLICT (slug) DO NOTHING;

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    subject TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_players (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    player_id BIGINT NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, game_id),
    CONSTRAINT uq_user_players_player UNIQUE (player_id)
);

CREATE INDEX IF NOT EXISTS ix_user_players_user_id
    ON user_players(user_id, game_id);

ALTER TABLE platform_account_links
    ADD COLUMN IF NOT EXISTS player_id BIGINT NULL REFERENCES players(id) ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS ix_platform_account_links_player_id
    ON platform_account_links(player_id, is_active, linked_at DESC)
    WHERE player_id IS NOT NULL;
