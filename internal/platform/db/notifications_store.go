package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sprocketbot/sprocket-v3/internal/domain/notifications"
)

type NotificationsStore struct {
	db *sql.DB
}

func NewNotificationsStore(db *sql.DB) *NotificationsStore {
	return &NotificationsStore{db: db}
}

func scanNotification(row interface {
	Scan(dest ...any) error
}) (notifications.PlayerNotification, error) {
	var n notifications.PlayerNotification
	var contextType sql.NullString
	var contextID sql.NullInt64
	var readAt sql.NullTime

	if err := row.Scan(
		&n.ID,
		&n.PlayerID,
		&n.Category,
		&contextType,
		&contextID,
		&n.Message,
		&readAt,
		&n.CreatedAt,
	); err != nil {
		return notifications.PlayerNotification{}, err
	}

	if contextType.Valid {
		n.ContextType = &contextType.String
	}
	if contextID.Valid {
		n.ContextID = &contextID.Int64
	}
	if readAt.Valid {
		n.ReadAt = &readAt.Time
	}

	return n, nil
}

func (s *NotificationsStore) CreateNotification(ctx context.Context, input notifications.CreateNotificationInput) (notifications.PlayerNotification, error) {
	const stmt = `
INSERT INTO player_notifications (player_id, category, context_type, context_id, message)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, player_id, category, context_type, context_id, message, read_at, created_at;`

	var contextType sql.NullString
	if input.ContextType != nil {
		contextType = sql.NullString{String: *input.ContextType, Valid: true}
	}

	var contextID sql.NullInt64
	if input.ContextID != nil {
		contextID = sql.NullInt64{Int64: *input.ContextID, Valid: true}
	}

	row := s.db.QueryRowContext(ctx, stmt,
		input.PlayerID,
		input.Category,
		contextType,
		contextID,
		input.Message,
	)

	return scanNotification(row)
}

func (s *NotificationsStore) ListNotifications(ctx context.Context, playerID int64, unreadOnly bool) ([]notifications.PlayerNotification, error) {
	stmt := `
SELECT id, player_id, category, context_type, context_id, message, read_at, created_at
FROM player_notifications
WHERE player_id = $1`

	if unreadOnly {
		stmt += ` AND read_at IS NULL`
	}

	stmt += `
ORDER BY (read_at IS NULL) DESC, created_at DESC;`

	rows, err := s.db.QueryContext(ctx, stmt, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]notifications.PlayerNotification, 0)
	for rows.Next() {
		n, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *NotificationsStore) MarkRead(ctx context.Context, input notifications.MarkReadInput) error {
	if input.MarkAll {
		const stmt = `
UPDATE player_notifications
SET read_at = NOW()
WHERE player_id = $1 AND read_at IS NULL;`

		_, err := s.db.ExecContext(ctx, stmt, input.PlayerID)
		return err
	}

	const stmt = `
UPDATE player_notifications
SET read_at = NOW()
WHERE player_id = $1 AND id = $2 AND read_at IS NULL;`

	for _, id := range input.IDs {
		if _, err := s.db.ExecContext(ctx, stmt, input.PlayerID, id); err != nil {
			return fmt.Errorf("marking notification %d as read: %w", id, err)
		}
	}

	return nil
}
