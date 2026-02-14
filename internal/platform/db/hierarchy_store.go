package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
)

type HierarchyStore struct {
	db *sql.DB
}

func NewHierarchyStore(db *sql.DB) *HierarchyStore {
	return &HierarchyStore{db: db}
}

func (s *HierarchyStore) CreateLeague(ctx context.Context, input hierarchy.CreateLeagueInput) (hierarchy.League, error) {
	if err := hierarchy.ValidateCreateLeagueInput(input); err != nil {
		return hierarchy.League{}, err
	}

	const stmt = `
INSERT INTO leagues(name, slug)
VALUES ($1, $2)
RETURNING id, name, slug, is_active, created_at;`
	var league hierarchy.League
	err := s.db.QueryRowContext(ctx, stmt, input.Name, input.Slug).Scan(
		&league.ID,
		&league.Name,
		&league.Slug,
		&league.IsActive,
		&league.CreatedAt,
	)
	if err != nil {
		return hierarchy.League{}, mapSQLError(err)
	}
	return league, nil
}

func (s *HierarchyStore) ListLeagues(ctx context.Context) ([]hierarchy.League, error) {
	const stmt = `
SELECT id, name, slug, is_active, created_at
FROM leagues
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leagues := make([]hierarchy.League, 0)
	for rows.Next() {
		var league hierarchy.League
		if err := rows.Scan(&league.ID, &league.Name, &league.Slug, &league.IsActive, &league.CreatedAt); err != nil {
			return nil, err
		}
		leagues = append(leagues, league)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return leagues, nil
}

func (s *HierarchyStore) CreateFranchise(ctx context.Context, input hierarchy.CreateFranchiseInput) (hierarchy.Franchise, error) {
	if err := hierarchy.ValidateCreateFranchiseInput(input); err != nil {
		return hierarchy.Franchise{}, err
	}

	const stmt = `
INSERT INTO franchises(league_id, name, slug)
VALUES ($1, $2, $3)
RETURNING id, league_id, name, slug, is_active, created_at;`
	var franchise hierarchy.Franchise
	err := s.db.QueryRowContext(ctx, stmt, input.LeagueID, input.Name, input.Slug).Scan(
		&franchise.ID,
		&franchise.LeagueID,
		&franchise.Name,
		&franchise.Slug,
		&franchise.IsActive,
		&franchise.CreatedAt,
	)
	if err != nil {
		return hierarchy.Franchise{}, mapSQLError(err)
	}
	return franchise, nil
}

func (s *HierarchyStore) ListFranchises(ctx context.Context) ([]hierarchy.Franchise, error) {
	const stmt = `
SELECT id, league_id, name, slug, is_active, created_at
FROM franchises
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	franchises := make([]hierarchy.Franchise, 0)
	for rows.Next() {
		var franchise hierarchy.Franchise
		if err := rows.Scan(&franchise.ID, &franchise.LeagueID, &franchise.Name, &franchise.Slug, &franchise.IsActive, &franchise.CreatedAt); err != nil {
			return nil, err
		}
		franchises = append(franchises, franchise)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return franchises, nil
}

func (s *HierarchyStore) CreateClub(ctx context.Context, input hierarchy.CreateClubInput) (hierarchy.Club, error) {
	if err := hierarchy.ValidateCreateClubInput(input); err != nil {
		return hierarchy.Club{}, err
	}

	const stmt = `
INSERT INTO clubs(franchise_id, name, slug)
VALUES ($1, $2, $3)
RETURNING id, franchise_id, name, slug, is_active, created_at;`
	var club hierarchy.Club
	err := s.db.QueryRowContext(ctx, stmt, input.FranchiseID, input.Name, input.Slug).Scan(
		&club.ID,
		&club.FranchiseID,
		&club.Name,
		&club.Slug,
		&club.IsActive,
		&club.CreatedAt,
	)
	if err != nil {
		return hierarchy.Club{}, mapSQLError(err)
	}
	return club, nil
}

func (s *HierarchyStore) ListClubs(ctx context.Context) ([]hierarchy.Club, error) {
	const stmt = `
SELECT id, franchise_id, name, slug, is_active, created_at
FROM clubs
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clubs := make([]hierarchy.Club, 0)
	for rows.Next() {
		var club hierarchy.Club
		if err := rows.Scan(&club.ID, &club.FranchiseID, &club.Name, &club.Slug, &club.IsActive, &club.CreatedAt); err != nil {
			return nil, err
		}
		clubs = append(clubs, club)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clubs, nil
}

func (s *HierarchyStore) CreateTeam(ctx context.Context, input hierarchy.CreateTeamInput) (hierarchy.Team, error) {
	if err := hierarchy.ValidateCreateTeamInput(input); err != nil {
		return hierarchy.Team{}, err
	}

	const stmt = `
INSERT INTO teams(club_id, name, slug)
VALUES ($1, $2, $3)
RETURNING id, club_id, name, slug, is_active, created_at;`
	var team hierarchy.Team
	err := s.db.QueryRowContext(ctx, stmt, input.ClubID, input.Name, input.Slug).Scan(
		&team.ID,
		&team.ClubID,
		&team.Name,
		&team.Slug,
		&team.IsActive,
		&team.CreatedAt,
	)
	if err != nil {
		return hierarchy.Team{}, mapSQLError(err)
	}
	return team, nil
}

func (s *HierarchyStore) ListTeams(ctx context.Context) ([]hierarchy.Team, error) {
	const stmt = `
SELECT id, club_id, name, slug, is_active, created_at
FROM teams
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]hierarchy.Team, 0)
	for rows.Next() {
		var team hierarchy.Team
		if err := rows.Scan(&team.ID, &team.ClubID, &team.Name, &team.Slug, &team.IsActive, &team.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return teams, nil
}

func (s *HierarchyStore) CreatePlayer(ctx context.Context, input hierarchy.CreatePlayerInput) (hierarchy.Player, error) {
	if err := hierarchy.ValidateCreatePlayerInput(input); err != nil {
		return hierarchy.Player{}, err
	}

	const stmt = `
INSERT INTO players(display_name, slug)
VALUES ($1, $2)
RETURNING id, display_name, slug, is_active, created_at;`
	var player hierarchy.Player
	err := s.db.QueryRowContext(ctx, stmt, input.DisplayName, input.Slug).Scan(
		&player.ID,
		&player.DisplayName,
		&player.Slug,
		&player.IsActive,
		&player.CreatedAt,
	)
	if err != nil {
		return hierarchy.Player{}, mapSQLError(err)
	}
	return player, nil
}

func (s *HierarchyStore) ListPlayers(ctx context.Context) ([]hierarchy.Player, error) {
	const stmt = `
SELECT id, display_name, slug, is_active, created_at
FROM players
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make([]hierarchy.Player, 0)
	for rows.Next() {
		var player hierarchy.Player
		if err := rows.Scan(&player.ID, &player.DisplayName, &player.Slug, &player.IsActive, &player.CreatedAt); err != nil {
			return nil, err
		}
		players = append(players, player)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return players, nil
}

func (s *HierarchyStore) CreateRosterMembership(ctx context.Context, input hierarchy.CreateRosterMembershipInput) (hierarchy.RosterMembership, error) {
	if err := hierarchy.ValidateCreateRosterMembershipInput(input); err != nil {
		return hierarchy.RosterMembership{}, err
	}

	const stmt = `
INSERT INTO roster_memberships(player_id, team_id)
VALUES ($1, $2)
RETURNING id, player_id, team_id, is_active, created_at;`
	var membership hierarchy.RosterMembership
	err := s.db.QueryRowContext(ctx, stmt, input.PlayerID, input.TeamID).Scan(
		&membership.ID,
		&membership.PlayerID,
		&membership.TeamID,
		&membership.IsActive,
		&membership.CreatedAt,
	)
	if err != nil {
		return hierarchy.RosterMembership{}, mapSQLError(err)
	}
	return membership, nil
}

func (s *HierarchyStore) ListRosterMemberships(ctx context.Context) ([]hierarchy.RosterMembership, error) {
	const stmt = `
SELECT id, player_id, team_id, is_active, created_at
FROM roster_memberships
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memberships := make([]hierarchy.RosterMembership, 0)
	for rows.Next() {
		var membership hierarchy.RosterMembership
		if err := rows.Scan(&membership.ID, &membership.PlayerID, &membership.TeamID, &membership.IsActive, &membership.CreatedAt); err != nil {
			return nil, err
		}
		memberships = append(memberships, membership)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return memberships, nil
}

func (s *HierarchyStore) CreateQueue(ctx context.Context, input hierarchy.CreateQueueInput) (hierarchy.Queue, error) {
	if err := hierarchy.ValidateCreateQueueInput(input); err != nil {
		return hierarchy.Queue{}, err
	}

	const stmt = `
INSERT INTO queues(name, slug)
VALUES ($1, $2)
RETURNING id, name, slug, is_active, created_at;`
	var queue hierarchy.Queue
	err := s.db.QueryRowContext(ctx, stmt, input.Name, input.Slug).Scan(
		&queue.ID,
		&queue.Name,
		&queue.Slug,
		&queue.IsActive,
		&queue.CreatedAt,
	)
	if err != nil {
		return hierarchy.Queue{}, mapSQLError(err)
	}
	return queue, nil
}

func (s *HierarchyStore) ListQueues(ctx context.Context) ([]hierarchy.Queue, error) {
	const stmt = `
SELECT id, name, slug, is_active, created_at
FROM queues
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	queues := make([]hierarchy.Queue, 0)
	for rows.Next() {
		var queue hierarchy.Queue
		if err := rows.Scan(&queue.ID, &queue.Name, &queue.Slug, &queue.IsActive, &queue.CreatedAt); err != nil {
			return nil, err
		}
		queues = append(queues, queue)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return queues, nil
}

func (s *HierarchyStore) EnqueueTeam(ctx context.Context, input hierarchy.EnqueueTeamInput) (hierarchy.QueueEntry, error) {
	if err := hierarchy.ValidateEnqueueTeamInput(input); err != nil {
		return hierarchy.QueueEntry{}, err
	}

	const stmt = `
INSERT INTO queue_entries(queue_id, team_id)
VALUES ($1, $2)
RETURNING id, queue_id, team_id, is_active, created_at, left_at;`
	var entry hierarchy.QueueEntry
	err := s.db.QueryRowContext(ctx, stmt, input.QueueID, input.TeamID).Scan(
		&entry.ID,
		&entry.QueueID,
		&entry.TeamID,
		&entry.IsActive,
		&entry.CreatedAt,
		&entry.LeftAt,
	)
	if err != nil {
		return hierarchy.QueueEntry{}, mapSQLError(err)
	}
	return entry, nil
}

func (s *HierarchyStore) LeaveQueue(ctx context.Context, input hierarchy.LeaveQueueInput) (hierarchy.QueueEntry, error) {
	if err := hierarchy.ValidateLeaveQueueInput(input); err != nil {
		return hierarchy.QueueEntry{}, err
	}

	const stmt = `
UPDATE queue_entries
SET is_active = FALSE, left_at = NOW()
WHERE queue_id = $1 AND team_id = $2 AND is_active = TRUE
RETURNING id, queue_id, team_id, is_active, created_at, left_at;`
	var entry hierarchy.QueueEntry
	err := s.db.QueryRowContext(ctx, stmt, input.QueueID, input.TeamID).Scan(
		&entry.ID,
		&entry.QueueID,
		&entry.TeamID,
		&entry.IsActive,
		&entry.CreatedAt,
		&entry.LeftAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.QueueEntry{}, fmt.Errorf("%w: active queue entry not found", hierarchy.ErrConflict)
		}
		return hierarchy.QueueEntry{}, mapSQLError(err)
	}
	return entry, nil
}

func (s *HierarchyStore) ListActiveQueueEntries(ctx context.Context) ([]hierarchy.QueueEntry, error) {
	const stmt = `
SELECT id, queue_id, team_id, is_active, created_at, left_at
FROM queue_entries
WHERE is_active = TRUE
ORDER BY queue_id ASC, created_at ASC, id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]hierarchy.QueueEntry, 0)
	for rows.Next() {
		var entry hierarchy.QueueEntry
		if err := rows.Scan(&entry.ID, &entry.QueueID, &entry.TeamID, &entry.IsActive, &entry.CreatedAt, &entry.LeftAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

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

func mapSQLError(err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.Code {
		case "23505":
			return fmt.Errorf("%w: %s", hierarchy.ErrConflict, pgError.ConstraintName)
		case "23503":
			return fmt.Errorf("%w: %s", hierarchy.ErrDependency, pgError.ConstraintName)
		case "23514":
			return fmt.Errorf("%w: %s", hierarchy.ErrInvalidInput, pgError.ConstraintName)
		}
	}
	return err
}
