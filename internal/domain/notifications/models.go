package notifications

import "time"

const (
	CategoryScrimPopped        = "scrim_popped"
	CategoryQueueBan           = "queue_ban"
	CategoryRatificationNeeded = "ratification_needed"
	CategorySkillGroupChange   = "skill_group_change"
	CategoryRosterChange       = "roster_change"
)

type PlayerNotification struct {
	ID          int64      `json:"id"`
	PlayerID    int64      `json:"playerId"`
	Category    string     `json:"category"`
	ContextType *string    `json:"contextType,omitempty"`
	ContextID   *int64     `json:"contextId,omitempty"`
	Message     string     `json:"message"`
	ReadAt      *time.Time `json:"readAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type CreateNotificationInput struct {
	PlayerID    int64   `json:"playerId"`
	Category    string  `json:"category"`
	ContextType *string `json:"contextType,omitempty"`
	ContextID   *int64  `json:"contextId,omitempty"`
	Message     string  `json:"message"`
}

type MarkReadInput struct {
	PlayerID int64   `json:"playerId"`
	IDs      []int64 `json:"ids"`
	MarkAll  bool    `json:"markAll"`
}
