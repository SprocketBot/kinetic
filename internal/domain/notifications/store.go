package notifications

import "context"

type Store interface {
	CreateNotification(ctx context.Context, input CreateNotificationInput) (PlayerNotification, error)
	// ListNotifications returns notifications for a player ordered by unread first, then newest first.
	ListNotifications(ctx context.Context, playerID int64, unreadOnly bool) ([]PlayerNotification, error)
	// MarkRead marks specific IDs (or all if input.MarkAll is true) as read for the given player.
	MarkRead(ctx context.Context, input MarkReadInput) error
}
