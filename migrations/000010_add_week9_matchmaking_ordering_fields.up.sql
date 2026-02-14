ALTER TABLE matchmaking_decisions
    ADD COLUMN IF NOT EXISTS wait_skew_seconds INT NOT NULL DEFAULT 0;

ALTER TABLE matchmaking_decisions
    ADD COLUMN IF NOT EXISTS home_team_rating INT NOT NULL DEFAULT 1000;

ALTER TABLE matchmaking_decisions
    ADD COLUMN IF NOT EXISTS away_team_rating INT NOT NULL DEFAULT 1000;

ALTER TABLE matchmaking_decisions
    ADD COLUMN IF NOT EXISTS ordering_strategy TEXT NOT NULL DEFAULT 'rating_spread_wait_skew_v1';
