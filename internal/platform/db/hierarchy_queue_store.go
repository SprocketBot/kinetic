package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
)

func (s *HierarchyStore) CreateQueue(ctx context.Context, input hierarchy.CreateQueueInput) (hierarchy.Queue, error) {
	if err := hierarchy.ValidateCreateQueueInput(input); err != nil {
		return hierarchy.Queue{}, err
	}

	const stmt = `
INSERT INTO queues(name, slug)
VALUES ($1, $2)
RETURNING id, name, slug, is_active, created_at;`
	var queue hierarchy.Queue
	err := s.db.QueryRowContext(ctx, stmt, input.Name, input.Slug).Scan(
		&queue.ID,
		&queue.Name,
		&queue.Slug,
		&queue.IsActive,
		&queue.CreatedAt,
	)
	if err != nil {
		return hierarchy.Queue{}, mapSQLError(err)
	}
	return queue, nil
}

func (s *HierarchyStore) ListQueues(ctx context.Context) ([]hierarchy.Queue, error) {
	const stmt = `
SELECT id, name, slug, is_active, created_at
FROM queues
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	queues := make([]hierarchy.Queue, 0)
	for rows.Next() {
		var queue hierarchy.Queue
		if err := rows.Scan(&queue.ID, &queue.Name, &queue.Slug, &queue.IsActive, &queue.CreatedAt); err != nil {
			return nil, err
		}
		queues = append(queues, queue)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return queues, nil
}

func (s *HierarchyStore) EnqueueTeam(ctx context.Context, input hierarchy.EnqueueTeamInput) (hierarchy.QueueEntry, error) {
	if err := hierarchy.ValidateEnqueueTeamInput(input); err != nil {
		return hierarchy.QueueEntry{}, err
	}

	const activeBanStmt = `
SELECT qb.player_id
FROM roster_memberships rm
JOIN queue_bans qb
	ON qb.player_id = rm.player_id
   AND qb.queue_id = $1
   AND qb.is_active = TRUE
WHERE rm.team_id = $2
  AND rm.is_active = TRUE
LIMIT 1;`
	var blockedPlayerID int64
	err := s.db.QueryRowContext(ctx, activeBanStmt, input.QueueID, input.TeamID).Scan(&blockedPlayerID)
	switch {
	case err == nil:
		return hierarchy.QueueEntry{}, fmt.Errorf("%w: team has player %d actively banned from this queue", hierarchy.ErrConflict, blockedPlayerID)
	case errors.Is(err, sql.ErrNoRows):
	default:
		return hierarchy.QueueEntry{}, err
	}

	const stmt = `
INSERT INTO queue_entries(queue_id, team_id)
VALUES ($1, $2)
RETURNING id, queue_id, team_id, is_active, expansion_stage, created_at, stage_advanced_at, left_at;`
	var entry hierarchy.QueueEntry
	err = s.db.QueryRowContext(ctx, stmt, input.QueueID, input.TeamID).Scan(
		&entry.ID,
		&entry.QueueID,
		&entry.TeamID,
		&entry.IsActive,
		&entry.Stage,
		&entry.CreatedAt,
		&entry.StageAt,
		&entry.LeftAt,
	)
	if err != nil {
		return hierarchy.QueueEntry{}, mapSQLError(err)
	}
	return entry, nil
}

func (s *HierarchyStore) LeaveQueue(ctx context.Context, input hierarchy.LeaveQueueInput) (hierarchy.QueueEntry, error) {
	if err := hierarchy.ValidateLeaveQueueInput(input); err != nil {
		return hierarchy.QueueEntry{}, err
	}

	const stmt = `
UPDATE queue_entries
SET is_active = FALSE, left_at = NOW()
WHERE queue_id = $1 AND team_id = $2 AND is_active = TRUE
RETURNING id, queue_id, team_id, is_active, expansion_stage, created_at, stage_advanced_at, left_at;`
	var entry hierarchy.QueueEntry
	err := s.db.QueryRowContext(ctx, stmt, input.QueueID, input.TeamID).Scan(
		&entry.ID,
		&entry.QueueID,
		&entry.TeamID,
		&entry.IsActive,
		&entry.Stage,
		&entry.CreatedAt,
		&entry.StageAt,
		&entry.LeftAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.QueueEntry{}, fmt.Errorf("%w: active queue entry not found", hierarchy.ErrConflict)
		}
		return hierarchy.QueueEntry{}, mapSQLError(err)
	}
	return entry, nil
}

func (s *HierarchyStore) BanPlayerFromQueue(ctx context.Context, input hierarchy.BanPlayerFromQueueInput) (hierarchy.QueueBan, error) {
	if err := hierarchy.ValidateBanPlayerFromQueueInput(input); err != nil {
		return hierarchy.QueueBan{}, err
	}

	const stmt = `
INSERT INTO queue_bans(queue_id, player_id, banned_by_actor, ban_reason)
VALUES ($1, $2, $3, $4)
RETURNING
	id,
	queue_id,
	player_id,
	banned_by_actor,
	ban_reason,
	is_active,
	banned_at,
	unbanned_by_actor,
	unban_reason,
	unbanned_at;`
	var ban hierarchy.QueueBan
	err := s.db.QueryRowContext(ctx, stmt, input.QueueID, input.PlayerID, input.Actor, input.Reason).Scan(
		&ban.ID,
		&ban.QueueID,
		&ban.PlayerID,
		&ban.BannedByActor,
		&ban.BanReason,
		&ban.IsActive,
		&ban.BannedAt,
		&ban.UnbannedByActor,
		&ban.UnbanReason,
		&ban.UnbannedAt,
	)
	if err != nil {
		return hierarchy.QueueBan{}, mapSQLError(err)
	}
	return ban, nil
}

func (s *HierarchyStore) UnbanPlayerFromQueue(ctx context.Context, input hierarchy.UnbanPlayerFromQueueInput) (hierarchy.QueueBan, error) {
	if err := hierarchy.ValidateUnbanPlayerFromQueueInput(input); err != nil {
		return hierarchy.QueueBan{}, err
	}

	const stmt = `
UPDATE queue_bans
SET
	is_active = FALSE,
	unbanned_by_actor = $3,
	unban_reason = $4,
	unbanned_at = NOW()
WHERE queue_id = $1
  AND player_id = $2
  AND is_active = TRUE
RETURNING
	id,
	queue_id,
	player_id,
	banned_by_actor,
	ban_reason,
	is_active,
	banned_at,
	unbanned_by_actor,
	unban_reason,
	unbanned_at;`
	var ban hierarchy.QueueBan
	err := s.db.QueryRowContext(ctx, stmt, input.QueueID, input.PlayerID, input.Actor, input.Reason).Scan(
		&ban.ID,
		&ban.QueueID,
		&ban.PlayerID,
		&ban.BannedByActor,
		&ban.BanReason,
		&ban.IsActive,
		&ban.BannedAt,
		&ban.UnbannedByActor,
		&ban.UnbanReason,
		&ban.UnbannedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.QueueBan{}, fmt.Errorf("%w: active queue ban not found", hierarchy.ErrConflict)
		}
		return hierarchy.QueueBan{}, mapSQLError(err)
	}
	return ban, nil
}

func (s *HierarchyStore) ListQueueBans(ctx context.Context) ([]hierarchy.QueueBan, error) {
	const stmt = `
SELECT
	id,
	queue_id,
	player_id,
	banned_by_actor,
	ban_reason,
	is_active,
	banned_at,
	unbanned_by_actor,
	unban_reason,
	unbanned_at
FROM queue_bans
ORDER BY id DESC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bans := make([]hierarchy.QueueBan, 0)
	for rows.Next() {
		var ban hierarchy.QueueBan
		if err := rows.Scan(
			&ban.ID,
			&ban.QueueID,
			&ban.PlayerID,
			&ban.BannedByActor,
			&ban.BanReason,
			&ban.IsActive,
			&ban.BannedAt,
			&ban.UnbannedByActor,
			&ban.UnbanReason,
			&ban.UnbannedAt,
		); err != nil {
			return nil, err
		}
		bans = append(bans, ban)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bans, nil
}

func (s *HierarchyStore) AdvanceQueueEntryStage(ctx context.Context, input hierarchy.AdvanceQueueEntryStageInput) (hierarchy.QueueEntry, error) {
	if err := hierarchy.ValidateAdvanceQueueEntryStageInput(input); err != nil {
		return hierarchy.QueueEntry{}, err
	}

	const stmt = `
UPDATE queue_entries
SET expansion_stage = $3, stage_advanced_at = NOW()
WHERE queue_id = $1
  AND team_id = $2
  AND is_active = TRUE
  AND $3 >= expansion_stage
RETURNING id, queue_id, team_id, is_active, expansion_stage, created_at, stage_advanced_at, left_at;`
	var entry hierarchy.QueueEntry
	err := s.db.QueryRowContext(ctx, stmt, input.QueueID, input.TeamID, input.Stage).Scan(
		&entry.ID,
		&entry.QueueID,
		&entry.TeamID,
		&entry.IsActive,
		&entry.Stage,
		&entry.CreatedAt,
		&entry.StageAt,
		&entry.LeftAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.QueueEntry{}, fmt.Errorf("%w: queue entry stage cannot decrease or entry not found", hierarchy.ErrConflict)
		}
		return hierarchy.QueueEntry{}, mapSQLError(err)
	}
	return entry, nil
}

func (s *HierarchyStore) ListActiveQueueEntries(ctx context.Context) ([]hierarchy.QueueEntry, error) {
	const stmt = `
SELECT id, queue_id, team_id, is_active, expansion_stage, created_at, stage_advanced_at, left_at
FROM queue_entries
WHERE is_active = TRUE
ORDER BY queue_id ASC, created_at ASC, id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]hierarchy.QueueEntry, 0)
	for rows.Next() {
		var entry hierarchy.QueueEntry
		if err := rows.Scan(&entry.ID, &entry.QueueID, &entry.TeamID, &entry.IsActive, &entry.Stage, &entry.CreatedAt, &entry.StageAt, &entry.LeftAt); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *HierarchyStore) ListActiveQueueBansByPlayerID(ctx context.Context, playerID int64) ([]hierarchy.QueueBan, error) {
	const stmt = `
SELECT id, queue_id, player_id, banned_by_actor, ban_reason, is_active, banned_at,
       unbanned_by_actor, unban_reason, unbanned_at
FROM queue_bans
WHERE player_id = $1 AND is_active = true
ORDER BY banned_at DESC;`
	rows, err := s.db.QueryContext(ctx, stmt, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bans := make([]hierarchy.QueueBan, 0)
	for rows.Next() {
		var ban hierarchy.QueueBan
		if err := rows.Scan(
			&ban.ID, &ban.QueueID, &ban.PlayerID, &ban.BannedByActor, &ban.BanReason,
			&ban.IsActive, &ban.BannedAt, &ban.UnbannedByActor, &ban.UnbanReason, &ban.UnbannedAt,
		); err != nil {
			return nil, err
		}
		bans = append(bans, ban)
	}
	return bans, rows.Err()
}

func (s *HierarchyStore) GetActiveQueueEntryByPlayerID(ctx context.Context, playerID int64) (*hierarchy.QueueEntry, error) {
	const stmt = `
SELECT qe.id, qe.queue_id, qe.team_id, qe.is_active, qe.stage, qe.created_at, qe.stage_at, qe.left_at
FROM queue_entries qe
JOIN roster_memberships rm ON rm.team_id = qe.team_id
    AND rm.player_id = $1 AND rm.is_active = true
WHERE qe.is_active = true AND qe.left_at IS NULL
LIMIT 1;`
	var qe hierarchy.QueueEntry
	err := s.db.QueryRowContext(ctx, stmt, playerID).Scan(
		&qe.ID, &qe.QueueID, &qe.TeamID, &qe.IsActive, &qe.Stage, &qe.CreatedAt, &qe.StageAt, &qe.LeftAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &qe, nil
}
