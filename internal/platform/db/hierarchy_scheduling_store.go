package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
)

func (s *HierarchyStore) CreateSeason(ctx context.Context, input hierarchy.CreateSeasonInput) (hierarchy.Season, error) {
	if err := hierarchy.ValidateCreateSeasonInput(input); err != nil {
		return hierarchy.Season{}, err
	}

	const stmt = `
INSERT INTO seasons(name, slug)
VALUES ($1, $2)
RETURNING id, name, slug, is_active, created_at;`
	var season hierarchy.Season
	err := s.db.QueryRowContext(ctx, stmt, input.Name, input.Slug).Scan(
		&season.ID,
		&season.Name,
		&season.Slug,
		&season.IsActive,
		&season.CreatedAt,
	)
	if err != nil {
		return hierarchy.Season{}, mapSQLError(err)
	}
	return season, nil
}

func (s *HierarchyStore) ListSeasons(ctx context.Context) ([]hierarchy.Season, error) {
	const stmt = `
SELECT id, name, slug, is_active, created_at
FROM seasons
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seasons := make([]hierarchy.Season, 0)
	for rows.Next() {
		var season hierarchy.Season
		if err := rows.Scan(&season.ID, &season.Name, &season.Slug, &season.IsActive, &season.CreatedAt); err != nil {
			return nil, err
		}
		seasons = append(seasons, season)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return seasons, nil
}

func (s *HierarchyStore) CreateScheduleGroup(ctx context.Context, input hierarchy.CreateScheduleGroupInput) (hierarchy.ScheduleGroup, error) {
	if err := hierarchy.ValidateCreateScheduleGroupInput(input); err != nil {
		return hierarchy.ScheduleGroup{}, err
	}

	const stmt = `
INSERT INTO schedule_groups(season_id, name, sequence)
VALUES ($1, $2, $3)
RETURNING id, season_id, name, sequence, is_active, created_at;`
	var group hierarchy.ScheduleGroup
	err := s.db.QueryRowContext(ctx, stmt, input.SeasonID, input.Name, input.Sequence).Scan(
		&group.ID,
		&group.SeasonID,
		&group.Name,
		&group.Sequence,
		&group.IsActive,
		&group.CreatedAt,
	)
	if err != nil {
		return hierarchy.ScheduleGroup{}, mapSQLError(err)
	}
	return group, nil
}

func (s *HierarchyStore) ListScheduleGroups(ctx context.Context) ([]hierarchy.ScheduleGroup, error) {
	const stmt = `
SELECT id, season_id, name, sequence, is_active, created_at
FROM schedule_groups
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]hierarchy.ScheduleGroup, 0)
	for rows.Next() {
		var group hierarchy.ScheduleGroup
		if err := rows.Scan(&group.ID, &group.SeasonID, &group.Name, &group.Sequence, &group.IsActive, &group.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

func (s *HierarchyStore) CreateFixture(ctx context.Context, input hierarchy.CreateFixtureInput) (hierarchy.Fixture, error) {
	if err := hierarchy.ValidateCreateFixtureInput(input); err != nil {
		return hierarchy.Fixture{}, err
	}

	const stmt = `
INSERT INTO fixtures(schedule_group_id, home_club_id, away_club_id)
VALUES ($1, $2, $3)
RETURNING id, schedule_group_id, home_club_id, away_club_id, is_active, created_at;`
	var fixture hierarchy.Fixture
	err := s.db.QueryRowContext(ctx, stmt, input.ScheduleGroupID, input.HomeClubID, input.AwayClubID).Scan(
		&fixture.ID,
		&fixture.ScheduleGroupID,
		&fixture.HomeClubID,
		&fixture.AwayClubID,
		&fixture.IsActive,
		&fixture.CreatedAt,
	)
	if err != nil {
		return hierarchy.Fixture{}, mapSQLError(err)
	}
	return fixture, nil
}

func (s *HierarchyStore) ListFixtures(ctx context.Context) ([]hierarchy.Fixture, error) {
	const stmt = `
SELECT id, schedule_group_id, home_club_id, away_club_id, is_active, created_at
FROM fixtures
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fixtures := make([]hierarchy.Fixture, 0)
	for rows.Next() {
		var fixture hierarchy.Fixture
		if err := rows.Scan(&fixture.ID, &fixture.ScheduleGroupID, &fixture.HomeClubID, &fixture.AwayClubID, &fixture.IsActive, &fixture.CreatedAt); err != nil {
			return nil, err
		}
		fixtures = append(fixtures, fixture)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return fixtures, nil
}

func (s *HierarchyStore) CreateMatch(ctx context.Context, input hierarchy.CreateMatchInput) (hierarchy.Match, error) {
	if err := hierarchy.ValidateCreateMatchInput(input); err != nil {
		return hierarchy.Match{}, err
	}

	const stmt = `
INSERT INTO matches(
	fixture_id, home_team_id, away_team_id, state, scheduled_for, home_time_ratified_at, away_time_ratified_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, fixture_id, home_team_id, away_team_id, state, scheduled_for, home_time_ratified_at, away_time_ratified_at, created_at;`
	var match hierarchy.Match
	err := s.db.QueryRowContext(
		ctx,
		stmt,
		input.FixtureID,
		input.HomeTeamID,
		input.AwayTeamID,
		input.State,
		input.ScheduledFor,
		input.HomeTimeRatifiedAt,
		input.AwayTimeRatifiedAt,
	).Scan(
		&match.ID,
		&match.FixtureID,
		&match.HomeTeamID,
		&match.AwayTeamID,
		&match.State,
		&match.ScheduledFor,
		&match.HomeTimeRatifiedAt,
		&match.AwayTimeRatifiedAt,
		&match.CreatedAt,
	)
	if err != nil {
		return hierarchy.Match{}, mapSQLError(err)
	}
	return match, nil
}

func (s *HierarchyStore) ListMatches(ctx context.Context) ([]hierarchy.Match, error) {
	const stmt = `
SELECT id, fixture_id, home_team_id, away_team_id, state, scheduled_for, home_time_ratified_at, away_time_ratified_at, created_at
FROM matches
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := make([]hierarchy.Match, 0)
	for rows.Next() {
		var match hierarchy.Match
		if err := rows.Scan(
			&match.ID,
			&match.FixtureID,
			&match.HomeTeamID,
			&match.AwayTeamID,
			&match.State,
			&match.ScheduledFor,
			&match.HomeTimeRatifiedAt,
			&match.AwayTimeRatifiedAt,
			&match.CreatedAt,
		); err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return matches, nil
}

func (s *HierarchyStore) GetFixture(ctx context.Context, fixtureID int64) (hierarchy.Fixture, error) {
	const stmt = `
SELECT id, schedule_group_id, home_club_id, away_club_id, is_active, created_at
FROM fixtures
WHERE id = $1;`
	var fixture hierarchy.Fixture
	err := s.db.QueryRowContext(ctx, stmt, fixtureID).Scan(
		&fixture.ID,
		&fixture.ScheduleGroupID,
		&fixture.HomeClubID,
		&fixture.AwayClubID,
		&fixture.IsActive,
		&fixture.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.Fixture{}, fmt.Errorf("%w: fixture not found", hierarchy.ErrDependency)
		}
		return hierarchy.Fixture{}, mapSQLError(err)
	}
	return fixture, nil
}
