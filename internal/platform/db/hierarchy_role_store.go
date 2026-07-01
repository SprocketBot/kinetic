package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
)

func (s *HierarchyStore) AssignRole(ctx context.Context, input hierarchy.AssignRoleInput) (hierarchy.RoleAssignment, error) {
	if err := hierarchy.ValidateAssignRoleInput(input); err != nil {
		return hierarchy.RoleAssignment{}, err
	}

	scope, err := s.resolveRoleScope(ctx, input.Role, input.FranchiseID, input.ClubID, input.TeamID)
	if err != nil {
		return hierarchy.RoleAssignment{}, err
	}

	allowed, err := s.canActorManageRole(ctx, input.ActorPlayerID, input.Role, scope, true)
	if err != nil {
		return hierarchy.RoleAssignment{}, err
	}
	if !allowed {
		return hierarchy.RoleAssignment{}, fmt.Errorf("%w: actor lacks scope authority for role assignment", hierarchy.ErrConflict)
	}

	const stmt = `
INSERT INTO role_assignments(
	player_id,
	role,
	franchise_id,
	club_id,
	team_id,
	assigned_by_actor_player_id,
	assign_reason
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING
	id,
	player_id,
	role,
	franchise_id,
	club_id,
	team_id,
	assigned_by_actor_player_id,
	assign_reason,
	is_active,
	assigned_at,
	revoked_by_actor_player_id,
	revoke_reason,
	revoked_at;`
	var assignment hierarchy.RoleAssignment
	err = s.db.QueryRowContext(
		ctx,
		stmt,
		input.TargetPlayerID,
		input.Role,
		scope.franchiseID,
		scope.clubID,
		scope.teamID,
		input.ActorPlayerID,
		input.Reason,
	).Scan(
		&assignment.ID,
		&assignment.PlayerID,
		&assignment.Role,
		&assignment.FranchiseID,
		&assignment.ClubID,
		&assignment.TeamID,
		&assignment.AssignedByActorPlayerID,
		&assignment.AssignReason,
		&assignment.IsActive,
		&assignment.AssignedAt,
		&assignment.RevokedByActorPlayerID,
		&assignment.RevokeReason,
		&assignment.RevokedAt,
	)
	if err != nil {
		return hierarchy.RoleAssignment{}, mapSQLError(err)
	}
	return assignment, nil
}

func (s *HierarchyStore) RevokeRole(ctx context.Context, input hierarchy.RevokeRoleInput) (hierarchy.RoleAssignment, error) {
	if err := hierarchy.ValidateRevokeRoleInput(input); err != nil {
		return hierarchy.RoleAssignment{}, err
	}

	const selectStmt = `
SELECT
	id,
	player_id,
	role,
	franchise_id,
	club_id,
	team_id,
	assigned_by_actor_player_id,
	assign_reason,
	is_active,
	assigned_at,
	revoked_by_actor_player_id,
	revoke_reason,
	revoked_at
FROM role_assignments
WHERE id = $1
  AND is_active = TRUE;`
	var existing hierarchy.RoleAssignment
	err := s.db.QueryRowContext(ctx, selectStmt, input.AssignmentID).Scan(
		&existing.ID,
		&existing.PlayerID,
		&existing.Role,
		&existing.FranchiseID,
		&existing.ClubID,
		&existing.TeamID,
		&existing.AssignedByActorPlayerID,
		&existing.AssignReason,
		&existing.IsActive,
		&existing.AssignedAt,
		&existing.RevokedByActorPlayerID,
		&existing.RevokeReason,
		&existing.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.RoleAssignment{}, fmt.Errorf("%w: active role assignment not found", hierarchy.ErrConflict)
		}
		return hierarchy.RoleAssignment{}, err
	}

	if existing.PlayerID != input.ActorPlayerID {
		allowed, err := s.canActorManageRole(ctx, input.ActorPlayerID, existing.Role, roleScope{
			franchiseID: existing.FranchiseID,
			clubID:      existing.ClubID,
			teamID:      existing.TeamID,
		}, false)
		if err != nil {
			return hierarchy.RoleAssignment{}, err
		}
		if !allowed {
			return hierarchy.RoleAssignment{}, fmt.Errorf("%w: actor lacks scope authority for role revoke", hierarchy.ErrConflict)
		}
	}

	const stmt = `
UPDATE role_assignments
SET
	is_active = FALSE,
	revoked_by_actor_player_id = $2,
	revoke_reason = $3,
	revoked_at = NOW()
WHERE id = $1
  AND is_active = TRUE
RETURNING
	id,
	player_id,
	role,
	franchise_id,
	club_id,
	team_id,
	assigned_by_actor_player_id,
	assign_reason,
	is_active,
	assigned_at,
	revoked_by_actor_player_id,
	revoke_reason,
	revoked_at;`
	var assignment hierarchy.RoleAssignment
	err = s.db.QueryRowContext(ctx, stmt, input.AssignmentID, input.ActorPlayerID, input.Reason).Scan(
		&assignment.ID,
		&assignment.PlayerID,
		&assignment.Role,
		&assignment.FranchiseID,
		&assignment.ClubID,
		&assignment.TeamID,
		&assignment.AssignedByActorPlayerID,
		&assignment.AssignReason,
		&assignment.IsActive,
		&assignment.AssignedAt,
		&assignment.RevokedByActorPlayerID,
		&assignment.RevokeReason,
		&assignment.RevokedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.RoleAssignment{}, fmt.Errorf("%w: active role assignment not found", hierarchy.ErrConflict)
		}
		return hierarchy.RoleAssignment{}, mapSQLError(err)
	}
	return assignment, nil
}

func (s *HierarchyStore) ListRoleAssignments(ctx context.Context) ([]hierarchy.RoleAssignment, error) {
	const stmt = `
SELECT
	id,
	player_id,
	role,
	franchise_id,
	club_id,
	team_id,
	assigned_by_actor_player_id,
	assign_reason,
	is_active,
	assigned_at,
	revoked_by_actor_player_id,
	revoke_reason,
	revoked_at
FROM role_assignments
ORDER BY id DESC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assignments := make([]hierarchy.RoleAssignment, 0)
	for rows.Next() {
		var assignment hierarchy.RoleAssignment
		if err := rows.Scan(
			&assignment.ID,
			&assignment.PlayerID,
			&assignment.Role,
			&assignment.FranchiseID,
			&assignment.ClubID,
			&assignment.TeamID,
			&assignment.AssignedByActorPlayerID,
			&assignment.AssignReason,
			&assignment.IsActive,
			&assignment.AssignedAt,
			&assignment.RevokedByActorPlayerID,
			&assignment.RevokeReason,
			&assignment.RevokedAt,
		); err != nil {
			return nil, err
		}
		assignments = append(assignments, assignment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return assignments, nil
}

type roleScope struct {
	franchiseID *int64
	clubID      *int64
	teamID      *int64
}

func (s *HierarchyStore) resolveRoleScope(
	ctx context.Context,
	role string,
	franchiseID *int64,
	clubID *int64,
	teamID *int64,
) (roleScope, error) {
	switch role {
	case "fm":
		if franchiseID == nil {
			return roleScope{}, fmt.Errorf("%w: franchise scope is required for fm role", hierarchy.ErrInvalidInput)
		}
		if err := s.ensureFranchiseExists(ctx, *franchiseID); err != nil {
			return roleScope{}, err
		}
		return roleScope{franchiseID: franchiseID}, nil
	case "gm", "agm":
		if clubID == nil {
			return roleScope{}, fmt.Errorf("%w: club scope is required for gm/agm role", hierarchy.ErrInvalidInput)
		}
		resolvedFranchiseID, err := s.resolveClubFranchise(ctx, *clubID)
		if err != nil {
			return roleScope{}, err
		}
		return roleScope{
			franchiseID: int64Ptr(resolvedFranchiseID),
			clubID:      clubID,
		}, nil
	case "captain":
		if teamID == nil {
			return roleScope{}, fmt.Errorf("%w: team scope is required for captain role", hierarchy.ErrInvalidInput)
		}
		resolvedClubID, resolvedFranchiseID, err := s.resolveTeamClubFranchise(ctx, *teamID)
		if err != nil {
			return roleScope{}, err
		}
		return roleScope{
			franchiseID: int64Ptr(resolvedFranchiseID),
			clubID:      int64Ptr(resolvedClubID),
			teamID:      teamID,
		}, nil
	default:
		return roleScope{}, fmt.Errorf("%w: unsupported role", hierarchy.ErrInvalidInput)
	}
}

func (s *HierarchyStore) ensureFranchiseExists(ctx context.Context, franchiseID int64) error {
	const stmt = `SELECT EXISTS(SELECT 1 FROM franchises WHERE id = $1);`
	var exists bool
	if err := s.db.QueryRowContext(ctx, stmt, franchiseID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("%w: franchise not found", hierarchy.ErrDependency)
	}
	return nil
}

func (s *HierarchyStore) resolveClubFranchise(ctx context.Context, clubID int64) (int64, error) {
	const stmt = `SELECT franchise_id FROM clubs WHERE id = $1;`
	var franchiseID int64
	err := s.db.QueryRowContext(ctx, stmt, clubID).Scan(&franchiseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%w: club not found", hierarchy.ErrDependency)
		}
		return 0, err
	}
	return franchiseID, nil
}

func (s *HierarchyStore) resolveTeamClubFranchise(ctx context.Context, teamID int64) (int64, int64, error) {
	const stmt = `
SELECT t.club_id, c.franchise_id
FROM teams t
JOIN clubs c ON c.id = t.club_id
WHERE t.id = $1;`
	var clubID int64
	var franchiseID int64
	err := s.db.QueryRowContext(ctx, stmt, teamID).Scan(&clubID, &franchiseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, fmt.Errorf("%w: team not found", hierarchy.ErrDependency)
		}
		return 0, 0, err
	}
	return clubID, franchiseID, nil
}

func (s *HierarchyStore) canActorManageRole(
	ctx context.Context,
	actorPlayerID int64,
	role string,
	scope roleScope,
	allowBootstrap bool,
) (bool, error) {
	if allowBootstrap && role == "fm" {
		count, err := s.countActiveRoleAssignments(ctx, "fm", scope.franchiseID, nil, nil)
		if err != nil {
			return false, err
		}
		if count == 0 {
			return true, nil
		}
	}

	hasFM, err := s.hasActiveRoleAssignment(ctx, actorPlayerID, "fm", scope.franchiseID, nil, nil)
	if err != nil {
		return false, err
	}
	if hasFM {
		return true, nil
	}

	switch role {
	case "fm", "gm":
		return false, nil
	case "agm":
		return s.hasActiveRoleAssignment(ctx, actorPlayerID, "gm", nil, scope.clubID, nil)
	case "captain":
		hasGM, err := s.hasActiveRoleAssignment(ctx, actorPlayerID, "gm", nil, scope.clubID, nil)
		if err != nil {
			return false, err
		}
		if hasGM {
			return true, nil
		}
		return s.hasActiveRoleAssignment(ctx, actorPlayerID, "agm", nil, scope.clubID, nil)
	default:
		return false, fmt.Errorf("%w: unsupported role", hierarchy.ErrInvalidInput)
	}
}

func (s *HierarchyStore) hasActiveRoleAssignment(
	ctx context.Context,
	playerID int64,
	role string,
	franchiseID *int64,
	clubID *int64,
	teamID *int64,
) (bool, error) {
	const stmt = `
SELECT EXISTS(
	SELECT 1
	FROM role_assignments
	WHERE player_id = $1
	  AND role = $2
	  AND is_active = TRUE
	  AND ($3::BIGINT IS NULL OR franchise_id = $3)
	  AND ($4::BIGINT IS NULL OR club_id = $4)
	  AND ($5::BIGINT IS NULL OR team_id = $5)
);`
	var exists bool
	err := s.db.QueryRowContext(ctx, stmt, playerID, role, franchiseID, clubID, teamID).Scan(&exists)
	return exists, err
}

func (s *HierarchyStore) countActiveRoleAssignments(
	ctx context.Context,
	role string,
	franchiseID *int64,
	clubID *int64,
	teamID *int64,
) (int64, error) {
	const stmt = `
SELECT COUNT(*)
FROM role_assignments
WHERE role = $1
  AND is_active = TRUE
  AND ($2::BIGINT IS NULL OR franchise_id = $2)
  AND ($3::BIGINT IS NULL OR club_id = $3)
  AND ($4::BIGINT IS NULL OR team_id = $4);`
	var count int64
	err := s.db.QueryRowContext(ctx, stmt, role, franchiseID, clubID, teamID).Scan(&count)
	return count, err
}

func int64Ptr(value int64) *int64 {
	return &value
}
