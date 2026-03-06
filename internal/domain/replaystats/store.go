package replaystats

import "context"

// Store provides access to replay stat data.
type Store interface {
	// ListStatsBySubmission returns player stat lines for all rounds linked to a result submission.
	ListStatsBySubmission(ctx context.Context, submissionID int64) ([]PlayerStatLine, error)
	// GetPlayerCareerStats returns aggregated career stats for a player.
	GetPlayerCareerStats(ctx context.Context, playerID int64) (PlayerCareerStats, error)
}
