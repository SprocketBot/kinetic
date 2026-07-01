package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
	"github.com/sprocketbot/sprocket-v3/internal/domain/notifications"
)

func (s *HierarchyStore) ListPlayerRatings(ctx context.Context) ([]hierarchy.PlayerRating, error) {
	const stmt = `
SELECT id, player_id, context_key, rating, uncertainty, matches_played, last_competed_at, is_active, updated_at
FROM player_ratings
WHERE is_active = TRUE
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ratings := make([]hierarchy.PlayerRating, 0)
	for rows.Next() {
		var rating hierarchy.PlayerRating
		if err := rows.Scan(
			&rating.ID,
			&rating.PlayerID,
			&rating.ContextKey,
			&rating.Rating,
			&rating.Uncertainty,
			&rating.MatchesPlayed,
			&rating.LastCompetedAt,
			&rating.IsActive,
			&rating.UpdatedAt,
		); err != nil {
			return nil, err
		}
		ratings = append(ratings, rating)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ratings, nil
}

func (s *HierarchyStore) AdjustPlayerRating(ctx context.Context, input hierarchy.AdjustPlayerRatingInput) (hierarchy.PlayerRating, error) {
	if err := hierarchy.ValidateAdjustPlayerRatingInput(input); err != nil {
		return hierarchy.PlayerRating{}, err
	}
	if input.ActorPlayerID == input.TargetPlayerID {
		return hierarchy.PlayerRating{}, fmt.Errorf("%w: actorPlayerId cannot adjust own rating", hierarchy.ErrConflict)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.PlayerRating{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	previousRating := defaultTeamRating
	previousUncertainty := int32(350)
	previousMatchesPlayed := int32(0)

	const selectPreviousStmt = `
SELECT rating, uncertainty, matches_played
FROM player_ratings
WHERE player_id = $1
  AND context_key = $2
  AND is_active = TRUE
FOR UPDATE;`
	err = tx.QueryRowContext(ctx, selectPreviousStmt, input.TargetPlayerID, input.ContextKey).Scan(
		&previousRating,
		&previousUncertainty,
		&previousMatchesPlayed,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return hierarchy.PlayerRating{}, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		previousRating = defaultTeamRating
		previousUncertainty = 350
		previousMatchesPlayed = 0
	}

	const upsertStmt = `
INSERT INTO player_ratings(player_id, context_key, rating, uncertainty, matches_played)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (player_id, context_key) WHERE is_active = TRUE
DO UPDATE SET
	rating = EXCLUDED.rating,
	uncertainty = EXCLUDED.uncertainty,
	matches_played = EXCLUDED.matches_played,
	updated_at = NOW()
RETURNING id, player_id, context_key, rating, uncertainty, matches_played, last_competed_at, is_active, updated_at;`
	var updated hierarchy.PlayerRating
	if err := tx.QueryRowContext(
		ctx,
		upsertStmt,
		input.TargetPlayerID,
		input.ContextKey,
		input.Rating,
		input.Uncertainty,
		input.MatchesPlayed,
	).Scan(
		&updated.ID,
		&updated.PlayerID,
		&updated.ContextKey,
		&updated.Rating,
		&updated.Uncertainty,
		&updated.MatchesPlayed,
		&updated.LastCompetedAt,
		&updated.IsActive,
		&updated.UpdatedAt,
	); err != nil {
		return hierarchy.PlayerRating{}, mapSQLError(err)
	}

	const auditStmt = `
INSERT INTO rating_adjustments(
	actor_player_id,
	target_player_id,
	context_key,
	previous_rating,
	new_rating,
	previous_uncertainty,
	new_uncertainty,
	previous_matches_played,
	new_matches_played,
	reason
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`
	if _, err := tx.ExecContext(
		ctx,
		auditStmt,
		input.ActorPlayerID,
		input.TargetPlayerID,
		input.ContextKey,
		previousRating,
		input.Rating,
		previousUncertainty,
		input.Uncertainty,
		previousMatchesPlayed,
		input.MatchesPlayed,
		input.Reason,
	); err != nil {
		return hierarchy.PlayerRating{}, mapSQLError(err)
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.PlayerRating{}, err
	}
	return updated, nil
}

func (s *HierarchyStore) ListRatingAdjustments(ctx context.Context) ([]hierarchy.RatingAdjustment, error) {
	const stmt = `
SELECT
	id,
	actor_player_id,
	target_player_id,
	context_key,
	previous_rating,
	new_rating,
	previous_uncertainty,
	new_uncertainty,
	previous_matches_played,
	new_matches_played,
	reason,
	created_at
FROM rating_adjustments
ORDER BY id DESC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	adjustments := make([]hierarchy.RatingAdjustment, 0)
	for rows.Next() {
		var adjustment hierarchy.RatingAdjustment
		if err := rows.Scan(
			&adjustment.ID,
			&adjustment.ActorPlayerID,
			&adjustment.TargetPlayerID,
			&adjustment.ContextKey,
			&adjustment.PreviousRating,
			&adjustment.NewRating,
			&adjustment.PreviousUncertainty,
			&adjustment.NewUncertainty,
			&adjustment.PreviousMatchesPlayed,
			&adjustment.NewMatchesPlayed,
			&adjustment.Reason,
			&adjustment.CreatedAt,
		); err != nil {
			return nil, err
		}
		adjustments = append(adjustments, adjustment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return adjustments, nil
}

func (s *HierarchyStore) ListMatchmakingDecisions(ctx context.Context) ([]hierarchy.MatchmakingDecision, error) {
	const stmt = `
SELECT
	id,
	scrim_id,
	queue_id,
	queue_wait_seconds,
	wait_skew_seconds,
	expansion_stage,
	rating_spread,
	home_team_rating,
	away_team_rating,
	cross_group,
	ordering_strategy,
	created_at
FROM matchmaking_decisions
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	decisions := make([]hierarchy.MatchmakingDecision, 0)
	for rows.Next() {
		var decision hierarchy.MatchmakingDecision
		if err := rows.Scan(
			&decision.ID,
			&decision.ScrimID,
			&decision.QueueID,
			&decision.QueueWaitSeconds,
			&decision.WaitSkewSeconds,
			&decision.ExpansionStage,
			&decision.RatingSpread,
			&decision.HomeTeamRating,
			&decision.AwayTeamRating,
			&decision.CrossGroup,
			&decision.OrderingStrategy,
			&decision.CreatedAt,
		); err != nil {
			return nil, err
		}
		decisions = append(decisions, decision)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return decisions, nil
}

func deriveTeamRatingForQueue(ctx context.Context, tx *sql.Tx, teamID int64, queueSlug string) (int32, error) {
	const stmt = `
SELECT COALESCE(ROUND(AVG(COALESCE(selected.rating, $3)))::INT, $3)::INT
FROM roster_memberships rm
LEFT JOIN LATERAL (
	SELECT pr.rating
	FROM player_ratings pr
	WHERE pr.player_id = rm.player_id
	  AND pr.is_active = TRUE
	  AND (pr.context_key = $2 OR pr.context_key = $4)
	ORDER BY
	  CASE WHEN pr.context_key = $2 THEN 0 ELSE 1 END,
	  pr.updated_at DESC,
	  pr.id DESC
	LIMIT 1
) selected ON TRUE
WHERE rm.team_id = $1
  AND rm.is_active = TRUE;`

	var rating int32
	if err := tx.QueryRowContext(ctx, stmt, teamID, queueSlug, defaultTeamRating, fallbackRatingContextGlobalKey).Scan(&rating); err != nil {
		return 0, err
	}
	return rating, nil
}

func isBetterPair(leftA, rightA, leftB, rightB promotionCandidate) bool {
	spreadA := abs32(leftA.teamRating - rightA.teamRating)
	spreadB := abs32(leftB.teamRating - rightB.teamRating)
	if spreadA != spreadB {
		return spreadA < spreadB
	}

	waitSkewA := abs32(leftA.queueWaitSeconds - rightA.queueWaitSeconds)
	waitSkewB := abs32(leftB.queueWaitSeconds - rightB.queueWaitSeconds)
	if waitSkewA != waitSkewB {
		return waitSkewA < waitSkewB
	}

	readyAtA := maxTime(leftA.entry.CreatedAt, rightA.entry.CreatedAt)
	readyAtB := maxTime(leftB.entry.CreatedAt, rightB.entry.CreatedAt)
	if !readyAtA.Equal(readyAtB) {
		return readyAtA.Before(readyAtB)
	}

	if leftA.entry.CreatedAt.UnixNano() != leftB.entry.CreatedAt.UnixNano() {
		return leftA.entry.CreatedAt.Before(leftB.entry.CreatedAt)
	}
	if leftA.entry.ID != leftB.entry.ID {
		return leftA.entry.ID < leftB.entry.ID
	}
	if rightA.entry.CreatedAt.UnixNano() != rightB.entry.CreatedAt.UnixNano() {
		return rightA.entry.CreatedAt.Before(rightB.entry.CreatedAt)
	}
	return rightA.entry.ID < rightB.entry.ID
}

// computeGlicko2 applies one Glicko-2 style match update.
// actualScore is 1.0 for a win, 0.0 for a loss.
// opponentRating is the average rating of the opposing team.
func computeGlicko2(snap playerRatingSnapshot, actualScore float64, opponentRating float64) (newRating, newUncertainty int32) {
	const minUncertainty = 50.0
	const uncertaintyDecayFactor = 0.95
	expected := 1.0 / (1.0 + math.Pow(10.0, (opponentRating-float64(snap.rating))/400.0))
	k := math.Max(16.0, math.Min(float64(snap.uncertainty)/3.0, 64.0))
	delta := k * (actualScore - expected)
	newR := math.Max(0, math.Min(9999, float64(snap.rating)+delta))
	newU := math.Max(minUncertainty, float64(snap.uncertainty)*uncertaintyDecayFactor)
	return int32(math.Round(newR)), int32(math.Round(newU))
}

// ratingRadiusForWait returns the maximum rating spread allowed for a pair
// based on the longer-waiting team's queue wait time (Theme 5D expansion windows).
func ratingRadiusForWait(waitSeconds int32) int32 {
	switch {
	case waitSeconds >= 600:
		return 400
	case waitSeconds >= 300:
		return 250
	case waitSeconds >= 120:
		return 150
	default:
		return 100
	}
}

// expansionStageForWait returns the discrete expansion stage (0–3) for a wait time.
func expansionStageForWait(waitSeconds int32) int32 {
	switch {
	case waitSeconds >= 600:
		return 3
	case waitSeconds >= 300:
		return 2
	case waitSeconds >= 120:
		return 1
	default:
		return 0
	}
}

func abs32(v int32) int32 {
	return int32(math.Abs(float64(v)))
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// applyRatingUpdatesInTx applies Glicko-2 rating updates for all participants in a
// ratified result submission within the given transaction.
func (s *HierarchyStore) applyRatingUpdatesInTx(ctx context.Context, tx *sql.Tx, sub hierarchy.ResultSubmission) error {
	// Resolve context_key: use the queue slug for scrims, fallback for other contexts.
	contextKey := fallbackRatingContextGlobalKey
	if sub.ContextType == "scrim" {
		const queueSlugStmt = `
SELECT q.slug FROM scrims sc JOIN queues q ON q.id = sc.queue_id WHERE sc.id = $1;`
		var slug string
		if err := tx.QueryRowContext(ctx, queueSlugStmt, sub.ContextID).Scan(&slug); err == nil {
			contextKey = slug
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}

	// Collect roster for a team.
	getRoster := func(teamID int64) ([]int64, error) {
		rows, err := tx.QueryContext(ctx,
			`SELECT player_id FROM roster_memberships WHERE team_id = $1 AND is_active = TRUE;`,
			teamID,
		)
		if err != nil {
			return nil, err
		}
		var ids []int64
		for rows.Next() {
			var pid int64
			if err := rows.Scan(&pid); err != nil {
				rows.Close()
				return nil, err
			}
			ids = append(ids, pid)
		}
		rows.Close()
		return ids, rows.Err()
	}

	winnerIDs, err := getRoster(sub.WinningTeamID)
	if err != nil {
		return err
	}
	loserIDs, err := getRoster(sub.LosingTeamID)
	if err != nil {
		return err
	}
	if len(winnerIDs) == 0 && len(loserIDs) == 0 {
		return nil
	}

	// Fetch current ratings for all participants.
	const ratingStmt = `
SELECT rating, uncertainty, matches_played
FROM player_ratings
WHERE player_id = $1 AND context_key = $2 AND is_active = TRUE;`
	snapshots := make(map[int64]playerRatingSnapshot)
	for _, pid := range append(winnerIDs, loserIDs...) {
		snap := playerRatingSnapshot{
			playerID:    pid,
			rating:      defaultTeamRating,
			uncertainty: 350,
		}
		err := tx.QueryRowContext(ctx, ratingStmt, pid, contextKey).Scan(
			&snap.rating, &snap.uncertainty, &snap.matchesPlayed,
		)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		snapshots[pid] = snap
	}

	// Compute average rating for a side.
	avgRating := func(ids []int64) float64 {
		if len(ids) == 0 {
			return float64(defaultTeamRating)
		}
		var sum int64
		for _, pid := range ids {
			sum += int64(snapshots[pid].rating)
		}
		return float64(sum) / float64(len(ids))
	}
	winnerAvg := avgRating(winnerIDs)
	loserAvg := avgRating(loserIDs)

	reason := "scrim_result"
	if sub.ContextType == "match" {
		reason = "league_result"
	}

	const upsertRatingStmt = `
INSERT INTO player_ratings(player_id, context_key, rating, uncertainty, matches_played, last_competed_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (player_id, context_key) WHERE is_active = TRUE
DO UPDATE SET
	rating = EXCLUDED.rating,
	uncertainty = EXCLUDED.uncertainty,
	matches_played = EXCLUDED.matches_played,
	last_competed_at = NOW(),
	updated_at = NOW();`

	const adjStmt = `
INSERT INTO rating_adjustments(
	actor_player_id, target_player_id, context_key,
	previous_rating, new_rating,
	previous_uncertainty, new_uncertainty,
	previous_matches_played, new_matches_played,
	reason
) VALUES (NULL, $1, $2, $3, $4, $5, $6, $7, $8, $9);`

	applyOne := func(pid int64, actualScore float64, opponentAvg float64, teamID int64) error {
		snap := snapshots[pid]
		newRating, newUncertainty := computeGlicko2(snap, actualScore, opponentAvg)
		newMatchesPlayed := snap.matchesPlayed + 1

		if _, err := tx.ExecContext(ctx, upsertRatingStmt,
			pid, contextKey, newRating, newUncertainty, newMatchesPlayed,
		); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, adjStmt,
			pid, contextKey,
			snap.rating, newRating,
			snap.uncertainty, newUncertainty,
			snap.matchesPlayed, newMatchesPlayed,
			reason,
		); err != nil {
			return err
		}
		return s.evaluateSkillGroupBoundaryInTx(ctx, tx, pid, teamID, snap.rating, newRating, sub.ID)
	}

	for _, pid := range winnerIDs {
		if err := applyOne(pid, 1.0, loserAvg, sub.WinningTeamID); err != nil {
			return err
		}
	}
	for _, pid := range loserIDs {
		if err := applyOne(pid, 0.0, winnerAvg, sub.LosingTeamID); err != nil {
			return err
		}
	}
	return nil
}

// evaluateSkillGroupBoundaryInTx checks whether a player's new rating crosses a
// skill group boundary and, if so, records the transition and notifies the player.
func (s *HierarchyStore) evaluateSkillGroupBoundaryInTx(ctx context.Context, tx *sql.Tx, playerID, teamID int64, oldRating, newRating int32, submissionID int64) error {
	// Derive league from team → club → franchise → league.
	const leagueStmt = `
SELECT f.league_id
FROM teams t
JOIN clubs c ON c.id = t.club_id
JOIN franchises f ON f.id = c.franchise_id
WHERE t.id = $1;`
	var leagueID int64
	if err := tx.QueryRowContext(ctx, leagueStmt, teamID).Scan(&leagueID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	// Find the skill group the player was in (based on their previous rating).
	const currentSgStmt = `
SELECT id, promotion_threshold, demotion_threshold
FROM skill_groups
WHERE league_id = $1 AND is_active = TRUE
ORDER BY
    CASE WHEN rating_floor <= $2 AND $2 <= rating_ceiling THEN 0 ELSE 1 END,
    display_order DESC
LIMIT 1;`
	var currentSgID int64
	var promotionThreshold, demotionThreshold sql.NullInt32
	if err := tx.QueryRowContext(ctx, currentSgStmt, leagueID, oldRating).Scan(
		&currentSgID, &promotionThreshold, &demotionThreshold,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	var targetSgID *int64
	var direction string
	if promotionThreshold.Valid && newRating >= promotionThreshold.Int32 {
		const nextUpStmt = `
SELECT id FROM skill_groups
WHERE league_id = $1 AND is_active = TRUE
  AND rating_floor > (SELECT rating_ceiling FROM skill_groups WHERE id = $2)
ORDER BY display_order ASC
LIMIT 1;`
		var nextID int64
		if err := tx.QueryRowContext(ctx, nextUpStmt, leagueID, currentSgID).Scan(&nextID); err == nil {
			targetSgID = &nextID
			direction = "promotion"
		}
	} else if demotionThreshold.Valid && newRating <= demotionThreshold.Int32 {
		const nextDownStmt = `
SELECT id FROM skill_groups
WHERE league_id = $1 AND is_active = TRUE
  AND rating_ceiling < (SELECT rating_floor FROM skill_groups WHERE id = $2)
ORDER BY display_order DESC
LIMIT 1;`
		var nextID int64
		if err := tx.QueryRowContext(ctx, nextDownStmt, leagueID, currentSgID).Scan(&nextID); err == nil {
			targetSgID = &nextID
			direction = "demotion"
		}
	}
	if targetSgID == nil {
		return nil
	}

	const transitionStmt = `
INSERT INTO skill_group_transitions(player_id, from_skill_group_id, to_skill_group_id, rating_at_transition, direction)
VALUES ($1, $2, $3, $4, $5);`
	if _, err := tx.ExecContext(ctx, transitionStmt, playerID, currentSgID, *targetSgID, newRating, direction); err != nil {
		return err
	}

	var msg string
	if direction == "promotion" {
		msg = fmt.Sprintf("Congratulations! Your rating of %d has earned you a promotion.", newRating)
	} else {
		msg = fmt.Sprintf("Your rating of %d has resulted in a demotion.", newRating)
	}
	_, err := tx.ExecContext(ctx,
		`INSERT INTO player_notifications(player_id, category, context_type, context_id, message) VALUES ($1, $2, 'result_submission', $3, $4);`,
		playerID, notifications.CategorySkillGroupChange, submissionID, msg,
	)
	return err
}
