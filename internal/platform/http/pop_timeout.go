package http

import (
	"context"
	"log/slog"
	"time"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
)

// startPopTimeoutWatcher launches a goroutine that cancels a popped scrim if it hasn't
// fully checked in within the given timeout. Default ban duration is 60 minutes.
func startPopTimeoutWatcher(scrimID int64, timeout time.Duration, store hierarchy.Store, logger *slog.Logger) {
	time.AfterFunc(timeout, func() {
		ctx := context.Background()
		if err := store.ExecutePopTimeout(ctx, hierarchy.ExecutePopTimeoutInput{
			ScrimID:                 scrimID,
			QueueBanDurationMinutes: 60,
		}); err != nil {
			logger.Error("pop timeout execution failed", "scrimId", scrimID, "error", err)
		}
	})
}
