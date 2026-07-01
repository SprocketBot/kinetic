package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
)

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

func (s *HierarchyStore) SetPlayerActive(ctx context.Context, input hierarchy.SetPlayerActiveInput) (hierarchy.Player, error) {
	const stmt = `
UPDATE players
SET is_active = $1
WHERE id = $2
RETURNING id, display_name, slug, is_active, created_at;`
	var player hierarchy.Player
	err := s.db.QueryRowContext(ctx, stmt, input.IsActive, input.PlayerID).Scan(
		&player.ID,
		&player.DisplayName,
		&player.Slug,
		&player.IsActive,
		&player.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.Player{}, fmt.Errorf("%w: player not found", hierarchy.ErrDependency)
		}
		return hierarchy.Player{}, mapSQLError(err)
	}
	return player, nil
}

func (s *HierarchyStore) GetActiveRosterMembershipByPlayerID(ctx context.Context, playerID int64) (*hierarchy.RosterMembership, error) {
	const stmt = `
SELECT id, player_id, team_id, is_active, created_at
FROM roster_memberships
WHERE player_id = $1 AND is_active = true
LIMIT 1;`
	var rm hierarchy.RosterMembership
	err := s.db.QueryRowContext(ctx, stmt, playerID).Scan(
		&rm.ID, &rm.PlayerID, &rm.TeamID, &rm.IsActive, &rm.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rm, nil
}
