package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
)

func (s *HierarchyStore) ResolveRoleScope(ctx context.Context, role string, franchiseID, clubID, teamID *int64) (hierarchy.HierarchyScope, error) {
	switch role {
	case "fm":
		if franchiseID == nil {
			return hierarchy.HierarchyScope{}, fmt.Errorf("%w: franchise scope is required", hierarchy.ErrInvalidInput)
		}
		return s.scopeForFranchise(ctx, *franchiseID)
	case "gm", "agm":
		if clubID == nil {
			return hierarchy.HierarchyScope{}, fmt.Errorf("%w: club scope is required", hierarchy.ErrInvalidInput)
		}
		return s.scopeForClub(ctx, *clubID)
	case "captain":
		if teamID == nil {
			return hierarchy.HierarchyScope{}, fmt.Errorf("%w: team scope is required", hierarchy.ErrInvalidInput)
		}
		return s.scopeForTeam(ctx, *teamID)
	default:
		return hierarchy.HierarchyScope{}, fmt.Errorf("%w: unsupported role", hierarchy.ErrInvalidInput)
	}
}

func (s *HierarchyStore) ResolveAssignmentScope(ctx context.Context, assignmentID int64) (hierarchy.HierarchyScope, error) {
	var role string
	var franchiseID, clubID, teamID *int64
	err := s.db.QueryRowContext(ctx, `SELECT role, franchise_id, club_id, team_id FROM role_assignments WHERE id = $1 AND is_active = TRUE`, assignmentID).Scan(&role, &franchiseID, &clubID, &teamID)
	if errors.Is(err, sql.ErrNoRows) {
		return hierarchy.HierarchyScope{}, fmt.Errorf("%w: active role assignment not found", hierarchy.ErrDependency)
	}
	if err != nil {
		return hierarchy.HierarchyScope{}, err
	}
	return s.ResolveRoleScope(ctx, role, franchiseID, clubID, teamID)
}

func (s *HierarchyStore) ResolveTeamScope(ctx context.Context, teamID int64) (hierarchy.HierarchyScope, error) {
	return s.scopeForTeam(ctx, teamID)
}

func (s *HierarchyStore) scopeForFranchise(ctx context.Context, franchiseID int64) (hierarchy.HierarchyScope, error) {
	var gameID int64
	err := s.db.QueryRowContext(ctx, `SELECT l.game_id FROM franchises f JOIN leagues l ON l.id = f.league_id WHERE f.id = $1`, franchiseID).Scan(&gameID)
	if errors.Is(err, sql.ErrNoRows) {
		return hierarchy.HierarchyScope{}, fmt.Errorf("%w: franchise not found", hierarchy.ErrDependency)
	}
	return hierarchy.HierarchyScope{GameID: gameID, FranchiseID: &franchiseID}, err
}

func (s *HierarchyStore) scopeForClub(ctx context.Context, clubID int64) (hierarchy.HierarchyScope, error) {
	var gameID, franchiseID int64
	err := s.db.QueryRowContext(ctx, `SELECT l.game_id, f.id FROM clubs c JOIN franchises f ON f.id = c.franchise_id JOIN leagues l ON l.id = f.league_id WHERE c.id = $1`, clubID).Scan(&gameID, &franchiseID)
	if errors.Is(err, sql.ErrNoRows) {
		return hierarchy.HierarchyScope{}, fmt.Errorf("%w: club not found", hierarchy.ErrDependency)
	}
	return hierarchy.HierarchyScope{GameID: gameID, FranchiseID: &franchiseID, ClubID: &clubID}, err
}

func (s *HierarchyStore) scopeForTeam(ctx context.Context, teamID int64) (hierarchy.HierarchyScope, error) {
	var gameID, franchiseID, clubID int64
	err := s.db.QueryRowContext(ctx, `SELECT l.game_id, f.id, c.id FROM teams t JOIN clubs c ON c.id = t.club_id JOIN franchises f ON f.id = c.franchise_id JOIN leagues l ON l.id = f.league_id WHERE t.id = $1`, teamID).Scan(&gameID, &franchiseID, &clubID)
	if errors.Is(err, sql.ErrNoRows) {
		return hierarchy.HierarchyScope{}, fmt.Errorf("%w: team not found", hierarchy.ErrDependency)
	}
	return hierarchy.HierarchyScope{GameID: gameID, FranchiseID: &franchiseID, ClubID: &clubID, TeamID: &teamID}, err
}
