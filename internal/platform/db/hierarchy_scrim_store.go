package db

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
	"github.com/kineticbot/kinetic-v3/internal/domain/notifications"
)

const scrimCols = `id, queue_id, home_team_id, away_team_id, state,
       lobby_name, lobby_password, popped_at, home_checked_in_at, away_checked_in_at,
       created_at, started_at, ended_at`

func scanScrim(scanner interface{ Scan(...any) error }) (hierarchy.Scrim, error) {
	var sc hierarchy.Scrim
	var lobbyName, lobbyPassword sql.NullString
	if err := scanner.Scan(
		&sc.ID, &sc.QueueID, &sc.HomeTeamID, &sc.AwayTeamID, &sc.State,
		&lobbyName, &lobbyPassword, &sc.PoppedAt, &sc.HomeCheckedInAt, &sc.AwayCheckedInAt,
		&sc.CreatedAt, &sc.StartedAt, &sc.EndedAt,
	); err != nil {
		return hierarchy.Scrim{}, err
	}
	if lobbyName.Valid {
		sc.LobbyName = &lobbyName.String
	}
	if lobbyPassword.Valid {
		sc.LobbyPassword = &lobbyPassword.String
	}
	return sc, nil
}

const lobbyCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateLobbyCredentials() (name, password string) {
	randBytes := func(n int) []byte {
		b := make([]byte, n)
		_, _ = rand.Read(b)
		return b
	}
	nameRaw := randBytes(6)
	var nb strings.Builder
	nb.WriteString("SPR-")
	for _, v := range nameRaw {
		nb.WriteByte(lobbyCharset[int(v)%len(lobbyCharset)])
	}
	passRaw := randBytes(8)
	var pb strings.Builder
	for _, v := range passRaw {
		pb.WriteByte(lobbyCharset[int(v)%len(lobbyCharset)])
	}
	return nb.String(), pb.String()
}

func (s *HierarchyStore) CreateScrim(ctx context.Context, input hierarchy.CreateScrimInput) (hierarchy.Scrim, error) {
	if err := hierarchy.ValidateCreateScrimInput(input); err != nil {
		return hierarchy.Scrim{}, err
	}

	const stmt = `
INSERT INTO scrims(queue_id, home_team_id, away_team_id, state)
VALUES ($1, $2, $3, $4)
RETURNING ` + scrimCols + `;`
	scrim, err := scanScrim(s.db.QueryRowContext(ctx, stmt, input.QueueID, input.HomeTeamID, input.AwayTeamID, input.State))
	if err != nil {
		return hierarchy.Scrim{}, mapSQLError(err)
	}
	return scrim, nil
}

func (s *HierarchyStore) UpdateScrimState(ctx context.Context, input hierarchy.UpdateScrimStateInput) (hierarchy.Scrim, error) {
	if err := hierarchy.ValidateUpdateScrimStateInput(input); err != nil {
		return hierarchy.Scrim{}, err
	}

	const stmt = `
UPDATE scrims
SET
	state = $2,
	started_at = CASE
		WHEN $2 = 'in_progress' AND started_at IS NULL THEN NOW()
		ELSE started_at
	END,
	ended_at = CASE
		WHEN $2 IN ('closed', 'voided', 'cancelled') THEN NOW()
		ELSE ended_at
	END
WHERE id = $1
  AND state <> $2
  AND (
		(state = 'created'     AND $2 IN ('popped', 'in_progress', 'voided'))
		OR (state = 'popped'   AND $2 IN ('in_progress', 'voided', 'cancelled'))
		OR (state = 'in_progress' AND $2 IN ('closed', 'voided'))
  )
RETURNING ` + scrimCols + `;`

	scrim, err := scanScrim(s.db.QueryRowContext(ctx, stmt, input.ScrimID, input.State))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.Scrim{}, fmt.Errorf("%w: scrim transition not allowed or scrim not found", hierarchy.ErrConflict)
		}
		return hierarchy.Scrim{}, mapSQLError(err)
	}
	return scrim, nil
}

func (s *HierarchyStore) ListScrims(ctx context.Context) ([]hierarchy.Scrim, error) {
	stmt := `SELECT ` + scrimCols + ` FROM scrims ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scrims := make([]hierarchy.Scrim, 0)
	for rows.Next() {
		sc, err := scanScrim(rows)
		if err != nil {
			return nil, err
		}
		scrims = append(scrims, sc)
	}
	return scrims, rows.Err()
}

func (s *HierarchyStore) PromoteQueueToScrim(ctx context.Context, input hierarchy.PromoteQueueToScrimInput) (hierarchy.Scrim, error) {
	if err := hierarchy.ValidatePromoteQueueToScrimInput(input); err != nil {
		return hierarchy.Scrim{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.Scrim{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const queueSlugStmt = `
SELECT slug
FROM queues
WHERE id = $1;`
	var queueSlug string
	if err := tx.QueryRowContext(ctx, queueSlugStmt, input.QueueID).Scan(&queueSlug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.Scrim{}, fmt.Errorf("%w: queue not found", hierarchy.ErrDependency)
		}
		return hierarchy.Scrim{}, err
	}

	const selectEntriesStmt = `
SELECT id, queue_id, team_id, is_active, expansion_stage, created_at, stage_advanced_at, left_at
FROM queue_entries
WHERE queue_id = $1 AND is_active = TRUE
ORDER BY created_at ASC, id ASC
FOR UPDATE;`
	rows, err := tx.QueryContext(ctx, selectEntriesStmt, input.QueueID)
	if err != nil {
		return hierarchy.Scrim{}, err
	}

	entries := make([]hierarchy.QueueEntry, 0)
	for rows.Next() {
		var entry hierarchy.QueueEntry
		if err := rows.Scan(&entry.ID, &entry.QueueID, &entry.TeamID, &entry.IsActive, &entry.Stage, &entry.CreatedAt, &entry.StageAt, &entry.LeftAt); err != nil {
			return hierarchy.Scrim{}, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return hierarchy.Scrim{}, err
	}
	if err := rows.Close(); err != nil {
		return hierarchy.Scrim{}, err
	}
	if len(entries) < 2 {
		return hierarchy.Scrim{}, fmt.Errorf("%w: insufficient active queue entries", hierarchy.ErrConflict)
	}

	candidates := make([]promotionCandidate, 0, len(entries))
	now := time.Now().UTC()
	for _, entry := range entries {
		teamRating, err := deriveTeamRatingForQueue(ctx, tx, entry.TeamID, queueSlug)
		if err != nil {
			return hierarchy.Scrim{}, err
		}
		waitSeconds := int32(now.Sub(entry.CreatedAt).Seconds())
		if waitSeconds < 0 {
			waitSeconds = 0
		}
		candidates = append(candidates, promotionCandidate{
			entry:            entry,
			teamRating:       teamRating,
			queueWaitSeconds: waitSeconds,
		})
	}

	bestLeft := -1
	bestRight := -1
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			maxWait := max32(candidates[i].queueWaitSeconds, candidates[j].queueWaitSeconds)
			radius := ratingRadiusForWait(maxWait)
			if abs32(candidates[i].teamRating-candidates[j].teamRating) > radius {
				continue
			}
			if bestLeft < 0 || isBetterPair(candidates[i], candidates[j], candidates[bestLeft], candidates[bestRight]) {
				bestLeft = i
				bestRight = j
			}
		}
	}
	if bestLeft < 0 {
		return hierarchy.Scrim{}, fmt.Errorf("%w: no valid pair within expansion radius", hierarchy.ErrConflict)
	}

	home := candidates[bestLeft]
	away := candidates[bestRight]
	ratingSpread := abs32(home.teamRating - away.teamRating)
	waitSkewSeconds := abs32(home.queueWaitSeconds - away.queueWaitSeconds)
	queueWaitSeconds := max32(home.queueWaitSeconds, away.queueWaitSeconds)
	expansionStage := expansionStageForWait(queueWaitSeconds)
	crossGroup := ratingSpread > 100

	lobbyName, lobbyPassword := generateLobbyCredentials()
	const createScrimStmt = `
INSERT INTO scrims(queue_id, home_team_id, away_team_id, state, lobby_name, lobby_password, popped_at)
VALUES ($1, $2, $3, 'popped', $4, $5, NOW())
RETURNING ` + scrimCols + `;`
	scrim, err := scanScrim(tx.QueryRowContext(ctx, createScrimStmt,
		input.QueueID, home.entry.TeamID, away.entry.TeamID, lobbyName, lobbyPassword,
	))
	if err != nil {
		return hierarchy.Scrim{}, mapSQLError(err)
	}

	const consumeEntriesStmt = `
UPDATE queue_entries
SET is_active = FALSE, left_at = NOW()
WHERE id = ANY($1) AND is_active = TRUE;`
	if _, err := tx.ExecContext(ctx, consumeEntriesStmt, []int64{home.entry.ID, away.entry.ID}); err != nil {
		return hierarchy.Scrim{}, err
	}

	const decisionStmt = `
INSERT INTO matchmaking_decisions(
	scrim_id,
	queue_id,
	queue_wait_seconds,
	wait_skew_seconds,
	expansion_stage,
	rating_spread,
	home_team_rating,
	away_team_rating,
	cross_group,
	ordering_strategy
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`
	if _, err := tx.ExecContext(
		ctx,
		decisionStmt,
		scrim.ID,
		input.QueueID,
		queueWaitSeconds,
		waitSkewSeconds,
		expansionStage,
		ratingSpread,
		home.teamRating,
		away.teamRating,
		crossGroup,
		matchmakingOrderingStrategyV1,
	); err != nil {
		return hierarchy.Scrim{}, err
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.Scrim{}, err
	}
	return scrim, nil
}

func (s *HierarchyStore) ProcessQueuePromotions(ctx context.Context, input hierarchy.ProcessQueuePromotionsInput) (hierarchy.ProcessQueuePromotionsResult, error) {
	if err := hierarchy.ValidateProcessQueuePromotionsInput(input); err != nil {
		return hierarchy.ProcessQueuePromotionsResult{}, err
	}
	startedAt := time.Now().UTC()

	queueIDs := make([]int64, 0)
	if input.QueueID > 0 {
		queueIDs = append(queueIDs, input.QueueID)
	} else {
		const queuesStmt = `
SELECT id
FROM queues
WHERE is_active = TRUE
ORDER BY id ASC;`
		rows, err := s.db.QueryContext(ctx, queuesStmt)
		if err != nil {
			return hierarchy.ProcessQueuePromotionsResult{}, err
		}
		defer rows.Close()

		for rows.Next() {
			var queueID int64
			if err := rows.Scan(&queueID); err != nil {
				return hierarchy.ProcessQueuePromotionsResult{}, err
			}
			queueIDs = append(queueIDs, queueID)
		}
		if err := rows.Err(); err != nil {
			return hierarchy.ProcessQueuePromotionsResult{}, err
		}
	}

	result := hierarchy.ProcessQueuePromotionsResult{
		PoppedScrimIDs: make([]int64, 0),
	}
	for _, queueID := range queueIDs {
		result.ProcessedQueues++
		for {
			scrim, err := s.PromoteQueueToScrim(ctx, hierarchy.PromoteQueueToScrimInput{QueueID: queueID})
			if err == nil {
				result.PromotionsCreated++
				result.PoppedScrimIDs = append(result.PoppedScrimIDs, scrim.ID)
				continue
			}
			if errors.Is(err, hierarchy.ErrConflict) || errors.Is(err, hierarchy.ErrDependency) {
				result.Conflicts++
				break
			}
			return hierarchy.ProcessQueuePromotionsResult{}, err
		}
	}

	const runStmt = `
INSERT INTO promotion_processing_runs(queue_id, processed_queues, promotions_created, conflicts, duration_ms)
VALUES ($1, $2, $3, $4, $5);`
	var queueID *int64
	if input.QueueID > 0 {
		queueID = &input.QueueID
	}
	durationMs := int32(time.Since(startedAt).Milliseconds())
	if durationMs < 0 {
		durationMs = 0
	}
	if _, err := s.db.ExecContext(
		ctx,
		runStmt,
		queueID,
		result.ProcessedQueues,
		result.PromotionsCreated,
		result.Conflicts,
		durationMs,
	); err != nil {
		return hierarchy.ProcessQueuePromotionsResult{}, err
	}

	return result, nil
}

func (s *HierarchyStore) ListPromotionProcessingRuns(ctx context.Context) ([]hierarchy.PromotionProcessingRun, error) {
	const stmt = `
SELECT id, queue_id, processed_queues, promotions_created, conflicts, duration_ms, created_at
FROM promotion_processing_runs
ORDER BY id DESC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := make([]hierarchy.PromotionProcessingRun, 0)
	for rows.Next() {
		var run hierarchy.PromotionProcessingRun
		if err := rows.Scan(
			&run.ID,
			&run.QueueID,
			&run.ProcessedQueues,
			&run.PromotionsCreated,
			&run.Conflicts,
			&run.DurationMs,
			&run.CreatedAt,
		); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return runs, nil
}

func (s *HierarchyStore) GetScrim(ctx context.Context, scrimID int64) (hierarchy.Scrim, error) {
	stmt := `SELECT ` + scrimCols + ` FROM scrims WHERE id = $1;`
	sc, err := scanScrim(s.db.QueryRowContext(ctx, stmt, scrimID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.Scrim{}, fmt.Errorf("%w: scrim not found", hierarchy.ErrDependency)
		}
		return hierarchy.Scrim{}, mapSQLError(err)
	}
	return sc, nil
}

func (s *HierarchyStore) GetActiveScrimByPlayerID(ctx context.Context, playerID int64) (*hierarchy.Scrim, error) {
	const stmt = `
SELECT s.id, s.queue_id, s.home_team_id, s.away_team_id, s.state,
       s.lobby_name, s.lobby_password, s.popped_at, s.home_checked_in_at, s.away_checked_in_at,
       s.created_at, s.started_at, s.ended_at
FROM scrims s
JOIN roster_memberships rm ON (rm.team_id = s.home_team_id OR rm.team_id = s.away_team_id)
    AND rm.player_id = $1 AND rm.is_active = true
WHERE s.state NOT IN ('closed', 'voided', 'cancelled')
LIMIT 1;`
	sc, err := scanScrim(s.db.QueryRowContext(ctx, stmt, playerID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sc, nil
}

func (s *HierarchyStore) CheckInScrim(ctx context.Context, input hierarchy.CheckInScrimInput) (hierarchy.Scrim, error) {
	const stmt = `
UPDATE scrims
SET
	home_checked_in_at = CASE WHEN home_team_id = $2 THEN COALESCE(home_checked_in_at, NOW()) ELSE home_checked_in_at END,
	away_checked_in_at = CASE WHEN away_team_id = $2 THEN COALESCE(away_checked_in_at, NOW()) ELSE away_checked_in_at END,
	state = CASE
		WHEN home_team_id = $2 AND away_checked_in_at IS NOT NULL THEN 'in_progress'
		WHEN away_team_id = $2 AND home_checked_in_at IS NOT NULL THEN 'in_progress'
		ELSE state
	END,
	started_at = CASE
		WHEN home_team_id = $2 AND away_checked_in_at IS NOT NULL AND started_at IS NULL THEN NOW()
		WHEN away_team_id = $2 AND home_checked_in_at IS NOT NULL AND started_at IS NULL THEN NOW()
		ELSE started_at
	END
WHERE id = $1
  AND state = 'popped'
  AND (home_team_id = $2 OR away_team_id = $2)
RETURNING ` + scrimCols + `;`

	sc, err := scanScrim(s.db.QueryRowContext(ctx, stmt, input.ScrimID, input.TeamID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.Scrim{}, fmt.Errorf("%w: scrim not found, not in popped state, or team is not a participant", hierarchy.ErrConflict)
		}
		return hierarchy.Scrim{}, mapSQLError(err)
	}
	return sc, nil
}

func (s *HierarchyStore) ExecutePopTimeout(ctx context.Context, input hierarchy.ExecutePopTimeoutInput) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Lock and verify scrim is still popped.
	const lockStmt = `
SELECT id, queue_id, home_team_id, away_team_id, home_checked_in_at, away_checked_in_at
FROM scrims WHERE id = $1 AND state = 'popped' FOR UPDATE;`
	var scrimID, queueID, homeTeamID, awayTeamID int64
	var homeCheckedInAt, awayCheckedInAt sql.NullTime
	err = tx.QueryRowContext(ctx, lockStmt, input.ScrimID).Scan(
		&scrimID, &queueID, &homeTeamID, &awayTeamID, &homeCheckedInAt, &awayCheckedInAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil // scrim already transitioned; no-op
	}
	if err != nil {
		return err
	}

	// Cancel the scrim.
	if _, err := tx.ExecContext(ctx, `UPDATE scrims SET state = 'cancelled', ended_at = NOW() WHERE id = $1`, scrimID); err != nil {
		return err
	}

	banDuration := time.Duration(input.QueueBanDurationMinutes) * time.Minute
	bannedAt := time.Now().UTC()
	banUntil := bannedAt.Add(banDuration)
	_ = banUntil // stored via ban_reason for now; future work can add expires_at to queue_bans

	// Determine which teams need bans.
	teamsToban := make([]int64, 0, 2)
	if !homeCheckedInAt.Valid {
		teamsToban = append(teamsToban, homeTeamID)
	}
	if !awayCheckedInAt.Valid {
		teamsToban = append(teamsToban, awayTeamID)
	}
	if len(teamsToban) == 0 {
		return tx.Commit() // both checked in somehow; just cancel
	}

	// Get queue slug for ban record.
	var queueSlug string
	if err := tx.QueryRowContext(ctx, `SELECT slug FROM queues WHERE id = $1`, queueID).Scan(&queueSlug); err != nil {
		queueSlug = "unknown"
	}

	for _, teamID := range teamsToban {
		// Get roster members for this team.
		rows, err := tx.QueryContext(ctx,
			`SELECT player_id FROM roster_memberships WHERE team_id = $1 AND is_active = true`, teamID)
		if err != nil {
			return err
		}
		playerIDs := make([]int64, 0)
		for rows.Next() {
			var pid int64
			if err := rows.Scan(&pid); err != nil {
				rows.Close()
				return err
			}
			playerIDs = append(playerIDs, pid)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}

		banReason := fmt.Sprintf("no-show: failed to check in for scrim %d within %d minutes", scrimID, input.QueueBanDurationMinutes)
		for _, playerID := range playerIDs {
			var banID int64
			err := tx.QueryRowContext(ctx,
				`INSERT INTO queue_bans(queue_id, player_id, banned_by_actor, ban_reason) VALUES($1, $2, 'system', $3) RETURNING id`,
				queueID, playerID, banReason,
			).Scan(&banID)
			if err != nil {
				return err
			}
			_, err = tx.ExecContext(ctx,
				`INSERT INTO player_notifications(player_id, category, context_type, context_id, message)
				 VALUES($1, $2, 'scrims', $3, $4)`,
				playerID, notifications.CategoryQueueBan, scrimID,
				fmt.Sprintf("You received a %d-minute queue ban for not checking in to your scrim.", input.QueueBanDurationMinutes),
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *HierarchyStore) GetScrimMetrics(ctx context.Context) (hierarchy.ScrimMetrics, error) {
	const stmt = `
SELECT
	(SELECT COUNT(DISTINCT rm.player_id)
	 FROM queue_entries qe
	 JOIN roster_memberships rm ON rm.team_id = qe.team_id AND rm.is_active = true
	 WHERE qe.is_active = true AND qe.left_at IS NULL)                                 AS players_queued,
	(SELECT COUNT(DISTINCT team_id)
	 FROM (
		 SELECT home_team_id AS team_id FROM scrims WHERE state IN ('popped', 'in_progress')
		 UNION ALL
		 SELECT away_team_id AS team_id FROM scrims WHERE state IN ('popped', 'in_progress')
	 ) t)                                                                               AS teams_in_scrim,
	(SELECT COUNT(*) FROM scrims WHERE state IN ('popped', 'in_progress'))             AS open_scrims,
	(SELECT COUNT(*) FROM scrims WHERE state = 'closed'
	 AND ended_at >= NOW() - INTERVAL '1 day')                                         AS scrims_closed_today,
	(SELECT COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY queue_wait_seconds), 0)
	 FROM matchmaking_decisions
	 WHERE created_at >= NOW() - INTERVAL '1 day')                                     AS avg_wait_seconds_p50;`

	var m hierarchy.ScrimMetrics
	err := s.db.QueryRowContext(ctx, stmt).Scan(
		&m.PlayersQueued,
		&m.TeamsInScrim,
		&m.OpenScrims,
		&m.ScrimsClosedToday,
		&m.AvgWaitSecondsP50,
	)
	if err != nil {
		return hierarchy.ScrimMetrics{}, err
	}
	return m, nil
}
