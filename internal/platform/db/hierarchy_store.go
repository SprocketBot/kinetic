package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
)

type HierarchyStore struct {
	db *sql.DB
}

const (
	defaultTeamRating              int32  = 1000
	matchmakingOrderingStrategyV1  string = "rating_spread_wait_skew_v1"
	fallbackRatingContextGlobalKey string = "scrim-3v3"
)

type promotionCandidate struct {
	entry            hierarchy.QueueEntry
	teamRating       int32
	queueWaitSeconds int32
}

type playerRatingSnapshot struct {
	playerID      int64
	rating        int32
	uncertainty   int32
	matchesPlayed int32
}

func NewHierarchyStore(db *sql.DB) *HierarchyStore {
	return &HierarchyStore{db: db}
}

func mapSQLError(err error) error {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		switch pgError.Code {
		case "23505":
			return fmt.Errorf("%w: %s", hierarchy.ErrConflict, pgError.ConstraintName)
		case "23503":
			return fmt.Errorf("%w: %s", hierarchy.ErrDependency, pgError.ConstraintName)
		case "23514":
			return fmt.Errorf("%w: %s", hierarchy.ErrInvalidInput, pgError.ConstraintName)
		}
	}
	return err
}
