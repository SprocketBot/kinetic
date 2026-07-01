package db

import (
	"context"

	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
)

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
