package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
)

func (s *HierarchyStore) UpsertUser(ctx context.Context, input hierarchy.UpsertUserInput) (hierarchy.User, error) {
	if err := hierarchy.ValidateUpsertUserInput(input); err != nil {
		return hierarchy.User{}, err
	}

	const stmt = `
INSERT INTO users(subject, display_name)
VALUES ($1, $2)
ON CONFLICT (subject) DO UPDATE
SET display_name = EXCLUDED.display_name, updated_at = NOW()
RETURNING id, subject, display_name, is_active, created_at, updated_at;`
	var user hierarchy.User
	err := s.db.QueryRowContext(ctx, stmt, input.Subject, input.DisplayName).Scan(
		&user.ID, &user.Subject, &user.DisplayName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return hierarchy.User{}, mapSQLError(err)
	}
	return user, nil
}

func (s *HierarchyStore) CreateGame(ctx context.Context, input hierarchy.CreateGameInput) (hierarchy.Game, error) {
	if err := hierarchy.ValidateCreateGameInput(input); err != nil {
		return hierarchy.Game{}, err
	}
	const stmt = `INSERT INTO games(name, slug) VALUES ($1, $2) RETURNING id, name, slug, is_active, created_at;`
	var game hierarchy.Game
	err := s.db.QueryRowContext(ctx, stmt, input.Name, input.Slug).Scan(&game.ID, &game.Name, &game.Slug, &game.IsActive, &game.CreatedAt)
	if err != nil {
		return hierarchy.Game{}, mapSQLError(err)
	}
	return game, nil
}

func (s *HierarchyStore) ListGames(ctx context.Context) ([]hierarchy.Game, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, slug, is_active, created_at FROM games ORDER BY id ASC;`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	games := make([]hierarchy.Game, 0)
	for rows.Next() {
		var game hierarchy.Game
		if err := rows.Scan(&game.ID, &game.Name, &game.Slug, &game.IsActive, &game.CreatedAt); err != nil {
			return nil, err
		}
		games = append(games, game)
	}
	return games, rows.Err()
}

func (s *HierarchyStore) CreateUserPlayer(ctx context.Context, input hierarchy.CreateUserPlayerInput) (hierarchy.UserPlayer, error) {
	if err := hierarchy.ValidateCreateUserPlayerInput(input); err != nil {
		return hierarchy.UserPlayer{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.UserPlayer{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var player hierarchy.Player
	err = tx.QueryRowContext(ctx, `
INSERT INTO players(display_name, slug)
VALUES ($1, $2)
RETURNING id, display_name, slug, is_active, created_at;`, input.DisplayName, input.Slug).Scan(
		&player.ID, &player.DisplayName, &player.Slug, &player.IsActive, &player.CreatedAt,
	)
	if err != nil {
		return hierarchy.UserPlayer{}, mapSQLError(err)
	}

	var result hierarchy.UserPlayer
	err = tx.QueryRowContext(ctx, `
INSERT INTO user_players(user_id, player_id, game_id)
VALUES ($1, $2, $3)
RETURNING user_id, player_id, game_id, created_at;`, input.UserID, player.ID, input.GameID).Scan(
		&result.UserID, &result.PlayerID, &result.GameID, &result.CreatedAt,
	)
	if err != nil {
		return hierarchy.UserPlayer{}, mapSQLError(err)
	}
	if err := tx.QueryRowContext(ctx, `SELECT id, name, slug, is_active, created_at FROM games WHERE id = $1`, input.GameID).Scan(
		&result.Game.ID, &result.Game.Name, &result.Game.Slug, &result.Game.IsActive, &result.Game.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.UserPlayer{}, fmt.Errorf("%w: game not found", hierarchy.ErrDependency)
		}
		return hierarchy.UserPlayer{}, err
	}
	result.Player = player
	if err := tx.Commit(); err != nil {
		return hierarchy.UserPlayer{}, err
	}
	return result, nil
}

func (s *HierarchyStore) ListUserPlayers(ctx context.Context, userID int64) ([]hierarchy.UserPlayer, error) {
	const stmt = `
SELECT up.user_id, up.player_id, up.game_id, up.created_at,
       p.id, p.display_name, p.slug, p.is_active, p.created_at,
       g.id, g.name, g.slug, g.is_active, g.created_at
FROM user_players up
JOIN players p ON p.id = up.player_id
JOIN games g ON g.id = up.game_id
WHERE up.user_id = $1
ORDER BY g.id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	players := make([]hierarchy.UserPlayer, 0)
	for rows.Next() {
		var result hierarchy.UserPlayer
		if err := rows.Scan(
			&result.UserID, &result.PlayerID, &result.GameID, &result.CreatedAt,
			&result.Player.ID, &result.Player.DisplayName, &result.Player.Slug, &result.Player.IsActive, &result.Player.CreatedAt,
			&result.Game.ID, &result.Game.Name, &result.Game.Slug, &result.Game.IsActive, &result.Game.CreatedAt,
		); err != nil {
			return nil, err
		}
		players = append(players, result)
	}
	return players, rows.Err()
}

func (s *HierarchyStore) UserOwnsPlayer(ctx context.Context, userID, playerID int64) (bool, error) {
	if userID <= 0 || playerID <= 0 {
		return false, fmt.Errorf("%w: userId and playerId must be greater than zero", hierarchy.ErrInvalidInput)
	}
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_players WHERE user_id = $1 AND player_id = $2)`, userID, playerID).Scan(&exists)
	return exists, err
}

func (s *HierarchyStore) GetUserPlayerIDForGame(ctx context.Context, userID, gameID int64) (int64, error) {
	var playerID int64
	err := s.db.QueryRowContext(ctx, `SELECT player_id FROM user_players WHERE user_id = $1 AND game_id = $2`, userID, gameID).Scan(&playerID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: user has no player for game", hierarchy.ErrDependency)
	}
	return playerID, err
}
