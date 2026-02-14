CREATE TABLE IF NOT EXISTS seasons (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_season_name UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS schedule_groups (
    id BIGSERIAL PRIMARY KEY,
    season_id BIGINT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    sequence INT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_schedule_group_season_sequence UNIQUE (season_id, sequence)
);

CREATE TABLE IF NOT EXISTS fixtures (
    id BIGSERIAL PRIMARY KEY,
    schedule_group_id BIGINT NOT NULL REFERENCES schedule_groups(id) ON DELETE CASCADE,
    home_club_id BIGINT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    away_club_id BIGINT NOT NULL REFERENCES clubs(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_fixture_clubs_differ CHECK (home_club_id <> away_club_id)
);

CREATE TABLE IF NOT EXISTS matches (
    id BIGSERIAL PRIMARY KEY,
    fixture_id BIGINT NOT NULL REFERENCES fixtures(id) ON DELETE CASCADE,
    home_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    away_team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    state TEXT NOT NULL,
    scheduled_for TIMESTAMPTZ NULL,
    home_time_ratified_at TIMESTAMPTZ NULL,
    away_time_ratified_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_match_teams_differ CHECK (home_team_id <> away_team_id),
    CONSTRAINT ck_match_state_allowed CHECK (state IN ('planned', 'ready')),
    CONSTRAINT ck_match_ready_requires_schedule_ratification CHECK (
        state <> 'ready' OR (
            scheduled_for IS NOT NULL AND
            home_time_ratified_at IS NOT NULL AND
            away_time_ratified_at IS NOT NULL AND
            home_time_ratified_at <= scheduled_for AND
            away_time_ratified_at <= scheduled_for
        )
    )
);
