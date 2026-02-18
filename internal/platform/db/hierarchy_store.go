package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

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

func NewHierarchyStore(db *sql.DB) *HierarchyStore {
	return &HierarchyStore{db: db}
}

func (s *HierarchyStore) CreateLeague(ctx context.Context, input hierarchy.CreateLeagueInput) (hierarchy.League, error) {
	if err := hierarchy.ValidateCreateLeagueInput(input); err != nil {
		return hierarchy.League{}, err
	}

	const stmt = `
INSERT INTO leagues(name, slug)
VALUES ($1, $2)
RETURNING id, name, slug, is_active, created_at;`
	var league hierarchy.League
	err := s.db.QueryRowContext(ctx, stmt, input.Name, input.Slug).Scan(
		&league.ID,
		&league.Name,
		&league.Slug,
		&league.IsActive,
		&league.CreatedAt,
	)
	if err != nil {
		return hierarchy.League{}, mapSQLError(err)
	}
	return league, nil
}

func (s *HierarchyStore) ListLeagues(ctx context.Context) ([]hierarchy.League, error) {
	const stmt = `
SELECT id, name, slug, is_active, created_at
FROM leagues
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	leagues := make([]hierarchy.League, 0)
	for rows.Next() {
		var league hierarchy.League
		if err := rows.Scan(&league.ID, &league.Name, &league.Slug, &league.IsActive, &league.CreatedAt); err != nil {
			return nil, err
		}
		leagues = append(leagues, league)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return leagues, nil
}

func (s *HierarchyStore) CreateFranchise(ctx context.Context, input hierarchy.CreateFranchiseInput) (hierarchy.Franchise, error) {
	if err := hierarchy.ValidateCreateFranchiseInput(input); err != nil {
		return hierarchy.Franchise{}, err
	}

	const stmt = `
INSERT INTO franchises(league_id, name, slug)
VALUES ($1, $2, $3)
RETURNING id, league_id, name, slug, is_active, created_at;`
	var franchise hierarchy.Franchise
	err := s.db.QueryRowContext(ctx, stmt, input.LeagueID, input.Name, input.Slug).Scan(
		&franchise.ID,
		&franchise.LeagueID,
		&franchise.Name,
		&franchise.Slug,
		&franchise.IsActive,
		&franchise.CreatedAt,
	)
	if err != nil {
		return hierarchy.Franchise{}, mapSQLError(err)
	}
	return franchise, nil
}

func (s *HierarchyStore) ListFranchises(ctx context.Context) ([]hierarchy.Franchise, error) {
	const stmt = `
SELECT id, league_id, name, slug, is_active, created_at
FROM franchises
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	franchises := make([]hierarchy.Franchise, 0)
	for rows.Next() {
		var franchise hierarchy.Franchise
		if err := rows.Scan(&franchise.ID, &franchise.LeagueID, &franchise.Name, &franchise.Slug, &franchise.IsActive, &franchise.CreatedAt); err != nil {
			return nil, err
		}
		franchises = append(franchises, franchise)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return franchises, nil
}

func (s *HierarchyStore) CreateClub(ctx context.Context, input hierarchy.CreateClubInput) (hierarchy.Club, error) {
	if err := hierarchy.ValidateCreateClubInput(input); err != nil {
		return hierarchy.Club{}, err
	}

	const stmt = `
INSERT INTO clubs(franchise_id, name, slug)
VALUES ($1, $2, $3)
RETURNING id, franchise_id, name, slug, is_active, created_at;`
	var club hierarchy.Club
	err := s.db.QueryRowContext(ctx, stmt, input.FranchiseID, input.Name, input.Slug).Scan(
		&club.ID,
		&club.FranchiseID,
		&club.Name,
		&club.Slug,
		&club.IsActive,
		&club.CreatedAt,
	)
	if err != nil {
		return hierarchy.Club{}, mapSQLError(err)
	}
	return club, nil
}

func (s *HierarchyStore) ListClubs(ctx context.Context) ([]hierarchy.Club, error) {
	const stmt = `
SELECT id, franchise_id, name, slug, is_active, created_at
FROM clubs
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clubs := make([]hierarchy.Club, 0)
	for rows.Next() {
		var club hierarchy.Club
		if err := rows.Scan(&club.ID, &club.FranchiseID, &club.Name, &club.Slug, &club.IsActive, &club.CreatedAt); err != nil {
			return nil, err
		}
		clubs = append(clubs, club)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clubs, nil
}

func (s *HierarchyStore) CreateTeam(ctx context.Context, input hierarchy.CreateTeamInput) (hierarchy.Team, error) {
	if err := hierarchy.ValidateCreateTeamInput(input); err != nil {
		return hierarchy.Team{}, err
	}

	const stmt = `
INSERT INTO teams(club_id, name, slug)
VALUES ($1, $2, $3)
RETURNING id, club_id, name, slug, is_active, created_at;`
	var team hierarchy.Team
	err := s.db.QueryRowContext(ctx, stmt, input.ClubID, input.Name, input.Slug).Scan(
		&team.ID,
		&team.ClubID,
		&team.Name,
		&team.Slug,
		&team.IsActive,
		&team.CreatedAt,
	)
	if err != nil {
		return hierarchy.Team{}, mapSQLError(err)
	}
	return team, nil
}

func (s *HierarchyStore) ListTeams(ctx context.Context) ([]hierarchy.Team, error) {
	const stmt = `
SELECT id, club_id, name, slug, is_active, created_at
FROM teams
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := make([]hierarchy.Team, 0)
	for rows.Next() {
		var team hierarchy.Team
		if err := rows.Scan(&team.ID, &team.ClubID, &team.Name, &team.Slug, &team.IsActive, &team.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return teams, nil
}

func (s *HierarchyStore) CreatePlayer(ctx context.Context, input hierarchy.CreatePlayerInput) (hierarchy.Player, error) {
	if err := hierarchy.ValidateCreatePlayerInput(input); err != nil {
		return hierarchy.Player{}, err
	}

	const stmt = `
INSERT INTO players(display_name, slug)
VALUES ($1, $2)
RETURNING id, display_name, slug, is_active, created_at;`
	var player hierarchy.Player
	err := s.db.QueryRowContext(ctx, stmt, input.DisplayName, input.Slug).Scan(
		&player.ID,
		&player.DisplayName,
		&player.Slug,
		&player.IsActive,
		&player.CreatedAt,
	)
	if err != nil {
		return hierarchy.Player{}, mapSQLError(err)
	}
	return player, nil
}

func (s *HierarchyStore) ListPlayers(ctx context.Context) ([]hierarchy.Player, error) {
	const stmt = `
SELECT id, display_name, slug, is_active, created_at
FROM players
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make([]hierarchy.Player, 0)
	for rows.Next() {
		var player hierarchy.Player
		if err := rows.Scan(&player.ID, &player.DisplayName, &player.Slug, &player.IsActive, &player.CreatedAt); err != nil {
			return nil, err
		}
		players = append(players, player)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return players, nil
}

func (s *HierarchyStore) CreateRosterMembership(ctx context.Context, input hierarchy.CreateRosterMembershipInput) (hierarchy.RosterMembership, error) {
	if err := hierarchy.ValidateCreateRosterMembershipInput(input); err != nil {
		return hierarchy.RosterMembership{}, err
	}

	const stmt = `
INSERT INTO roster_memberships(player_id, team_id)
VALUES ($1, $2)
RETURNING id, player_id, team_id, is_active, created_at;`
	var membership hierarchy.RosterMembership
	err := s.db.QueryRowContext(ctx, stmt, input.PlayerID, input.TeamID).Scan(
		&membership.ID,
		&membership.PlayerID,
		&membership.TeamID,
		&membership.IsActive,
		&membership.CreatedAt,
	)
	if err != nil {
		return hierarchy.RosterMembership{}, mapSQLError(err)
	}
	return membership, nil
}

func (s *HierarchyStore) ListRosterMemberships(ctx context.Context) ([]hierarchy.RosterMembership, error) {
	const stmt = `
SELECT id, player_id, team_id, is_active, created_at
FROM roster_memberships
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memberships := make([]hierarchy.RosterMembership, 0)
	for rows.Next() {
		var membership hierarchy.RosterMembership
		if err := rows.Scan(&membership.ID, &membership.PlayerID, &membership.TeamID, &membership.IsActive, &membership.CreatedAt); err != nil {
			return nil, err
		}
		memberships = append(memberships, membership)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return memberships, nil
}

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

	const stmt = `
INSERT INTO queue_entries(queue_id, team_id)
VALUES ($1, $2)
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

func (s *HierarchyStore) CreateScrim(ctx context.Context, input hierarchy.CreateScrimInput) (hierarchy.Scrim, error) {
	if err := hierarchy.ValidateCreateScrimInput(input); err != nil {
		return hierarchy.Scrim{}, err
	}

	const stmt = `
INSERT INTO scrims(queue_id, home_team_id, away_team_id, state)
VALUES ($1, $2, $3, $4)
RETURNING id, queue_id, home_team_id, away_team_id, state, created_at, started_at, ended_at;`
	var scrim hierarchy.Scrim
	err := s.db.QueryRowContext(ctx, stmt, input.QueueID, input.HomeTeamID, input.AwayTeamID, input.State).Scan(
		&scrim.ID,
		&scrim.QueueID,
		&scrim.HomeTeamID,
		&scrim.AwayTeamID,
		&scrim.State,
		&scrim.CreatedAt,
		&scrim.StartedAt,
		&scrim.EndedAt,
	)
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
		WHEN $2 IN ('closed', 'voided') THEN NOW()
		ELSE ended_at
	END
WHERE id = $1
  AND state <> $2
  AND (
		(state = 'created' AND $2 IN ('in_progress', 'voided'))
		OR (state = 'in_progress' AND $2 IN ('closed', 'voided'))
  )
RETURNING id, queue_id, home_team_id, away_team_id, state, created_at, started_at, ended_at;`

	var scrim hierarchy.Scrim
	err := s.db.QueryRowContext(ctx, stmt, input.ScrimID, input.State).Scan(
		&scrim.ID,
		&scrim.QueueID,
		&scrim.HomeTeamID,
		&scrim.AwayTeamID,
		&scrim.State,
		&scrim.CreatedAt,
		&scrim.StartedAt,
		&scrim.EndedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.Scrim{}, fmt.Errorf("%w: scrim transition not allowed or scrim not found", hierarchy.ErrConflict)
		}
		return hierarchy.Scrim{}, mapSQLError(err)
	}

	return scrim, nil
}

func (s *HierarchyStore) ListScrims(ctx context.Context) ([]hierarchy.Scrim, error) {
	const stmt = `
SELECT id, queue_id, home_team_id, away_team_id, state, created_at, started_at, ended_at
FROM scrims
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scrims := make([]hierarchy.Scrim, 0)
	for rows.Next() {
		var scrim hierarchy.Scrim
		if err := rows.Scan(
			&scrim.ID,
			&scrim.QueueID,
			&scrim.HomeTeamID,
			&scrim.AwayTeamID,
			&scrim.State,
			&scrim.CreatedAt,
			&scrim.StartedAt,
			&scrim.EndedAt,
		); err != nil {
			return nil, err
		}
		scrims = append(scrims, scrim)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return scrims, nil
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

	bestLeft := 0
	bestRight := 1
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if isBetterPair(candidates[i], candidates[j], candidates[bestLeft], candidates[bestRight]) {
				bestLeft = i
				bestRight = j
			}
		}
	}

	home := candidates[bestLeft]
	away := candidates[bestRight]
	ratingSpread := abs32(home.teamRating - away.teamRating)
	waitSkewSeconds := abs32(home.queueWaitSeconds - away.queueWaitSeconds)
	queueWaitSeconds := max32(home.queueWaitSeconds, away.queueWaitSeconds)
	expansionStage := max32(home.entry.Stage, away.entry.Stage)

	const createScrimStmt = `
INSERT INTO scrims(queue_id, home_team_id, away_team_id, state)
VALUES ($1, $2, $3, 'created')
RETURNING id, queue_id, home_team_id, away_team_id, state, created_at, started_at, ended_at;`
	var scrim hierarchy.Scrim
	err = tx.QueryRowContext(ctx, createScrimStmt, input.QueueID, home.entry.TeamID, away.entry.TeamID).Scan(
		&scrim.ID,
		&scrim.QueueID,
		&scrim.HomeTeamID,
		&scrim.AwayTeamID,
		&scrim.State,
		&scrim.CreatedAt,
		&scrim.StartedAt,
		&scrim.EndedAt,
	)
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
		false,
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

	result := hierarchy.ProcessQueuePromotionsResult{}
	for _, queueID := range queueIDs {
		result.ProcessedQueues++
		for {
			_, err := s.PromoteQueueToScrim(ctx, hierarchy.PromoteQueueToScrimInput{QueueID: queueID})
			if err == nil {
				result.PromotionsCreated++
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

func (s *HierarchyStore) CreateResultSubmission(ctx context.Context, input hierarchy.CreateResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	if err := hierarchy.ValidateCreateResultSubmissionInput(input); err != nil {
		return hierarchy.ResultSubmission{}, err
	}

	homeTeamID, awayTeamID, err := resolveContextTeams(ctx, s.db, input.ContextType, input.ContextID)
	if err != nil {
		return hierarchy.ResultSubmission{}, err
	}
	if input.SubmittedByTeamID != homeTeamID && input.SubmittedByTeamID != awayTeamID {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: submittedByTeamId must be a context participant", hierarchy.ErrConflict)
	}
	if (input.WinningTeamID != homeTeamID && input.WinningTeamID != awayTeamID) ||
		(input.LosingTeamID != homeTeamID && input.LosingTeamID != awayTeamID) {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: winning and losing teams must match context participants", hierarchy.ErrConflict)
	}

	payload := input.PayloadJSON
	if len(payload) == 0 {
		payload = []byte(`{}`)
	}

	const stmt = `
INSERT INTO result_submissions(
	context_type,
	context_id,
	submitted_by_team_id,
	home_team_id,
	away_team_id,
	winning_team_id,
	losing_team_id,
	payload_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING
	id,
	context_type,
	context_id,
	submitted_by_team_id,
	home_team_id,
	away_team_id,
	winning_team_id,
	losing_team_id,
	state,
	payload_json,
	home_ratified_at,
	away_ratified_at,
	rejected_by_team_id,
	rejection_reason,
	rejected_at,
	created_at;`

	var submission hierarchy.ResultSubmission
	err = s.db.QueryRowContext(
		ctx,
		stmt,
		input.ContextType,
		input.ContextID,
		input.SubmittedByTeamID,
		homeTeamID,
		awayTeamID,
		input.WinningTeamID,
		input.LosingTeamID,
		payload,
	).Scan(
		&submission.ID,
		&submission.ContextType,
		&submission.ContextID,
		&submission.SubmittedByTeamID,
		&submission.HomeTeamID,
		&submission.AwayTeamID,
		&submission.WinningTeamID,
		&submission.LosingTeamID,
		&submission.State,
		&submission.PayloadJSON,
		&submission.HomeRatifiedAt,
		&submission.AwayRatifiedAt,
		&submission.RejectedByTeamID,
		&submission.RejectionReason,
		&submission.RejectedAt,
		&submission.CreatedAt,
	)
	if err != nil {
		return hierarchy.ResultSubmission{}, mapSQLError(err)
	}
	return submission, nil
}

func (s *HierarchyStore) ListResultSubmissions(ctx context.Context) ([]hierarchy.ResultSubmission, error) {
	const stmt = `
SELECT
	id,
	context_type,
	context_id,
	submitted_by_team_id,
	home_team_id,
	away_team_id,
	winning_team_id,
	losing_team_id,
	state,
	payload_json,
	home_ratified_at,
	away_ratified_at,
	rejected_by_team_id,
	rejection_reason,
	rejected_at,
	created_at
FROM result_submissions
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	submissions := make([]hierarchy.ResultSubmission, 0)
	for rows.Next() {
		var submission hierarchy.ResultSubmission
		if err := rows.Scan(
			&submission.ID,
			&submission.ContextType,
			&submission.ContextID,
			&submission.SubmittedByTeamID,
			&submission.HomeTeamID,
			&submission.AwayTeamID,
			&submission.WinningTeamID,
			&submission.LosingTeamID,
			&submission.State,
			&submission.PayloadJSON,
			&submission.HomeRatifiedAt,
			&submission.AwayRatifiedAt,
			&submission.RejectedByTeamID,
			&submission.RejectionReason,
			&submission.RejectedAt,
			&submission.CreatedAt,
		); err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return submissions, nil
}

func (s *HierarchyStore) OverrideResultSubmission(ctx context.Context, input hierarchy.OverrideResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	if err := hierarchy.ValidateOverrideResultSubmissionInput(input); err != nil {
		return hierarchy.ResultSubmission{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.ResultSubmission{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const lockStmt = `
SELECT
	home_team_id,
	away_team_id,
	winning_team_id,
	losing_team_id,
	state
FROM result_submissions
WHERE id = $1
FOR UPDATE;`
	var homeTeamID, awayTeamID int64
	var previousWinningTeamID, previousLosingTeamID int64
	var previousState string
	if err := tx.QueryRowContext(ctx, lockStmt, input.SubmissionID).Scan(
		&homeTeamID,
		&awayTeamID,
		&previousWinningTeamID,
		&previousLosingTeamID,
		&previousState,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.ResultSubmission{}, fmt.Errorf("%w: submission not found", hierarchy.ErrConflict)
		}
		return hierarchy.ResultSubmission{}, err
	}

	if (input.WinningTeamID != homeTeamID && input.WinningTeamID != awayTeamID) ||
		(input.LosingTeamID != homeTeamID && input.LosingTeamID != awayTeamID) {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: winning and losing teams must match submission participants", hierarchy.ErrConflict)
	}

	const overrideStmt = `
UPDATE result_submissions
SET
	winning_team_id = $2,
	losing_team_id = $3,
	state = 'ratified',
	home_ratified_at = COALESCE(home_ratified_at, NOW()),
	away_ratified_at = COALESCE(away_ratified_at, NOW()),
	rejected_by_team_id = NULL,
	rejection_reason = NULL,
	rejected_at = NULL
WHERE id = $1
RETURNING
	id,
	context_type,
	context_id,
	submitted_by_team_id,
	home_team_id,
	away_team_id,
	winning_team_id,
	losing_team_id,
	state,
	payload_json,
	home_ratified_at,
	away_ratified_at,
	rejected_by_team_id,
	rejection_reason,
	rejected_at,
	created_at;`
	var submission hierarchy.ResultSubmission
	if err := tx.QueryRowContext(ctx, overrideStmt, input.SubmissionID, input.WinningTeamID, input.LosingTeamID).Scan(
		&submission.ID,
		&submission.ContextType,
		&submission.ContextID,
		&submission.SubmittedByTeamID,
		&submission.HomeTeamID,
		&submission.AwayTeamID,
		&submission.WinningTeamID,
		&submission.LosingTeamID,
		&submission.State,
		&submission.PayloadJSON,
		&submission.HomeRatifiedAt,
		&submission.AwayRatifiedAt,
		&submission.RejectedByTeamID,
		&submission.RejectionReason,
		&submission.RejectedAt,
		&submission.CreatedAt,
	); err != nil {
		return hierarchy.ResultSubmission{}, mapSQLError(err)
	}

	const auditStmt = `
INSERT INTO result_overrides(
	submission_id,
	actor,
	reason,
	previous_winning_team_id,
	previous_losing_team_id,
	new_winning_team_id,
	new_losing_team_id,
	previous_state,
	new_state
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`
	if _, err := tx.ExecContext(
		ctx,
		auditStmt,
		input.SubmissionID,
		input.Actor,
		input.Reason,
		previousWinningTeamID,
		previousLosingTeamID,
		input.WinningTeamID,
		input.LosingTeamID,
		previousState,
		"ratified",
	); err != nil {
		return hierarchy.ResultSubmission{}, mapSQLError(err)
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ResultSubmission{}, err
	}
	return submission, nil
}

func (s *HierarchyStore) ListResultOverrides(ctx context.Context) ([]hierarchy.ResultOverride, error) {
	const stmt = `
SELECT
	id,
	submission_id,
	actor,
	reason,
	previous_winning_team_id,
	previous_losing_team_id,
	new_winning_team_id,
	new_losing_team_id,
	previous_state,
	new_state,
	created_at
FROM result_overrides
ORDER BY id DESC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	overrides := make([]hierarchy.ResultOverride, 0)
	for rows.Next() {
		var override hierarchy.ResultOverride
		if err := rows.Scan(
			&override.ID,
			&override.SubmissionID,
			&override.Actor,
			&override.Reason,
			&override.PreviousWinningTeamID,
			&override.PreviousLosingTeamID,
			&override.NewWinningTeamID,
			&override.NewLosingTeamID,
			&override.PreviousState,
			&override.NewState,
			&override.CreatedAt,
		); err != nil {
			return nil, err
		}
		overrides = append(overrides, override)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return overrides, nil
}

func (s *HierarchyStore) RatifyResultSubmission(ctx context.Context, input hierarchy.RatifyResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	if err := hierarchy.ValidateRatifyResultSubmissionInput(input); err != nil {
		return hierarchy.ResultSubmission{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.ResultSubmission{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const lockStmt = `
SELECT home_team_id, away_team_id, state, home_ratified_at, away_ratified_at
FROM result_submissions
WHERE id = $1
FOR UPDATE;`
	var homeTeamID, awayTeamID int64
	var state string
	var homeRatifiedAt, awayRatifiedAt *time.Time
	if err := tx.QueryRowContext(ctx, lockStmt, input.SubmissionID).Scan(
		&homeTeamID,
		&awayTeamID,
		&state,
		&homeRatifiedAt,
		&awayRatifiedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.ResultSubmission{}, fmt.Errorf("%w: submission not found", hierarchy.ErrConflict)
		}
		return hierarchy.ResultSubmission{}, err
	}
	if state != "pending" {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: only pending submissions can be ratified", hierarchy.ErrConflict)
	}
	if input.TeamID != homeTeamID && input.TeamID != awayTeamID {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: ratifying team must be a participant", hierarchy.ErrConflict)
	}
	if (input.TeamID == homeTeamID && homeRatifiedAt != nil) || (input.TeamID == awayTeamID && awayRatifiedAt != nil) {
		return hierarchy.ResultSubmission{}, fmt.Errorf("%w: team already ratified", hierarchy.ErrConflict)
	}

	const ratifyStmt = `
UPDATE result_submissions
SET
	home_ratified_at = CASE
		WHEN $2 = home_team_id AND home_ratified_at IS NULL THEN NOW()
		ELSE home_ratified_at
	END,
	away_ratified_at = CASE
		WHEN $2 = away_team_id AND away_ratified_at IS NULL THEN NOW()
		ELSE away_ratified_at
	END,
	state = CASE
		WHEN
			(CASE WHEN $2 = home_team_id AND home_ratified_at IS NULL THEN NOW() ELSE home_ratified_at END) IS NOT NULL
			AND
			(CASE WHEN $2 = away_team_id AND away_ratified_at IS NULL THEN NOW() ELSE away_ratified_at END) IS NOT NULL
		THEN 'ratified'
		ELSE state
	END
WHERE id = $1
RETURNING
	id,
	context_type,
	context_id,
	submitted_by_team_id,
	home_team_id,
	away_team_id,
	winning_team_id,
	losing_team_id,
	state,
	payload_json,
	home_ratified_at,
	away_ratified_at,
	rejected_by_team_id,
	rejection_reason,
	rejected_at,
	created_at;`
	var submission hierarchy.ResultSubmission
	if err := tx.QueryRowContext(ctx, ratifyStmt, input.SubmissionID, input.TeamID).Scan(
		&submission.ID,
		&submission.ContextType,
		&submission.ContextID,
		&submission.SubmittedByTeamID,
		&submission.HomeTeamID,
		&submission.AwayTeamID,
		&submission.WinningTeamID,
		&submission.LosingTeamID,
		&submission.State,
		&submission.PayloadJSON,
		&submission.HomeRatifiedAt,
		&submission.AwayRatifiedAt,
		&submission.RejectedByTeamID,
		&submission.RejectionReason,
		&submission.RejectedAt,
		&submission.CreatedAt,
	); err != nil {
		return hierarchy.ResultSubmission{}, mapSQLError(err)
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ResultSubmission{}, err
	}
	return submission, nil
}

func (s *HierarchyStore) RejectResultSubmission(ctx context.Context, input hierarchy.RejectResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	if err := hierarchy.ValidateRejectResultSubmissionInput(input); err != nil {
		return hierarchy.ResultSubmission{}, err
	}

	const stmt = `
UPDATE result_submissions
SET
	state = 'rejected',
	rejected_by_team_id = $2,
	rejection_reason = $3,
	rejected_at = NOW()
WHERE id = $1
  AND state = 'pending'
  AND ($2 = home_team_id OR $2 = away_team_id)
RETURNING
	id,
	context_type,
	context_id,
	submitted_by_team_id,
	home_team_id,
	away_team_id,
	winning_team_id,
	losing_team_id,
	state,
	payload_json,
	home_ratified_at,
	away_ratified_at,
	rejected_by_team_id,
	rejection_reason,
	rejected_at,
	created_at;`
	var submission hierarchy.ResultSubmission
	err := s.db.QueryRowContext(ctx, stmt, input.SubmissionID, input.TeamID, input.Reason).Scan(
		&submission.ID,
		&submission.ContextType,
		&submission.ContextID,
		&submission.SubmittedByTeamID,
		&submission.HomeTeamID,
		&submission.AwayTeamID,
		&submission.WinningTeamID,
		&submission.LosingTeamID,
		&submission.State,
		&submission.PayloadJSON,
		&submission.HomeRatifiedAt,
		&submission.AwayRatifiedAt,
		&submission.RejectedByTeamID,
		&submission.RejectionReason,
		&submission.RejectedAt,
		&submission.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.ResultSubmission{}, fmt.Errorf("%w: submission reject not allowed or not found", hierarchy.ErrConflict)
		}
		return hierarchy.ResultSubmission{}, mapSQLError(err)
	}
	return submission, nil
}

func (s *HierarchyStore) IngestReplayEvidence(ctx context.Context, input hierarchy.IngestReplayEvidenceInput) (hierarchy.ReplayIngestionResult, error) {
	if err := hierarchy.ValidateIngestReplayEvidenceInput(input); err != nil {
		return hierarchy.ReplayIngestionResult{}, err
	}

	homeTeamID, awayTeamID, err := resolveContextTeams(ctx, s.db, input.ContextType, input.ContextID)
	if err != nil {
		return hierarchy.ReplayIngestionResult{}, err
	}
	if input.SubmittedByTeamID != homeTeamID && input.SubmittedByTeamID != awayTeamID {
		return hierarchy.ReplayIngestionResult{}, fmt.Errorf("%w: submittedByTeamId must be a context participant", hierarchy.ErrConflict)
	}

	contentHash := sha256.Sum256([]byte(input.ReplayBody))
	replaySHA256 := hex.EncodeToString(contentHash[:])
	contentSizeBytes := int64(len(input.ReplayBody))
	output := input.ParseOutputJSON
	if len(output) == 0 {
		output = []byte(`{}`)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.ReplayIngestionResult{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	result := hierarchy.ReplayIngestionResult{}

	const findEvidenceStmt = `
SELECT id, context_type, context_id, submitted_by_team_id, replay_sha256, content_size_bytes, storage_ref, state, created_at
FROM replay_evidence
WHERE replay_sha256 = $1
FOR UPDATE;`
	err = tx.QueryRowContext(ctx, findEvidenceStmt, replaySHA256).Scan(
		&result.Evidence.ID,
		&result.Evidence.ContextType,
		&result.Evidence.ContextID,
		&result.Evidence.SubmittedByTeamID,
		&result.Evidence.ReplaySHA256,
		&result.Evidence.ContentSizeBytes,
		&result.Evidence.StorageRef,
		&result.Evidence.State,
		&result.Evidence.CreatedAt,
	)
	switch {
	case err == nil:
		if result.Evidence.ContextType != input.ContextType || result.Evidence.ContextID != input.ContextID {
			return hierarchy.ReplayIngestionResult{}, fmt.Errorf("%w: replay hash already attached to another context", hierarchy.ErrConflict)
		}
		result.Duplicate = true
	case errors.Is(err, sql.ErrNoRows):
		const createEvidenceStmt = `
INSERT INTO replay_evidence(
	context_type,
	context_id,
	submitted_by_team_id,
	replay_sha256,
	content_size_bytes,
	storage_ref,
	state
)
VALUES ($1, $2, $3, $4, $5, $6, 'parsed')
RETURNING id, context_type, context_id, submitted_by_team_id, replay_sha256, content_size_bytes, storage_ref, state, created_at;`
		if err := tx.QueryRowContext(
			ctx,
			createEvidenceStmt,
			input.ContextType,
			input.ContextID,
			input.SubmittedByTeamID,
			replaySHA256,
			contentSizeBytes,
			fmt.Sprintf("inline-sha256:%s", replaySHA256),
		).Scan(
			&result.Evidence.ID,
			&result.Evidence.ContextType,
			&result.Evidence.ContextID,
			&result.Evidence.SubmittedByTeamID,
			&result.Evidence.ReplaySHA256,
			&result.Evidence.ContentSizeBytes,
			&result.Evidence.StorageRef,
			&result.Evidence.State,
			&result.Evidence.CreatedAt,
		); err != nil {
			return hierarchy.ReplayIngestionResult{}, mapSQLError(err)
		}
	default:
		return hierarchy.ReplayIngestionResult{}, err
	}

	const createParseRunStmt = `
INSERT INTO replay_parse_runs(
	replay_evidence_id,
	parser_name,
	parser_version,
	parser_config_digest,
	status,
	output_json
)
VALUES ($1, $2, $3, $4, 'parsed', $5)
RETURNING id, replay_evidence_id, parser_name, parser_version, parser_config_digest, status, output_json, created_at;`
	if err := tx.QueryRowContext(
		ctx,
		createParseRunStmt,
		result.Evidence.ID,
		input.ParserName,
		input.ParserVersion,
		input.ParserConfigDigest,
		output,
	).Scan(
		&result.ParseRun.ID,
		&result.ParseRun.ReplayEvidenceID,
		&result.ParseRun.ParserName,
		&result.ParseRun.ParserVersion,
		&result.ParseRun.ParserConfigDigest,
		&result.ParseRun.Status,
		&result.ParseRun.OutputJSON,
		&result.ParseRun.CreatedAt,
	); err != nil {
		return hierarchy.ReplayIngestionResult{}, mapSQLError(err)
	}

	if input.ResultSubmissionID != nil {
		const verifySubmissionStmt = `
SELECT context_type, context_id, home_team_id, away_team_id
FROM result_submissions
WHERE id = $1;`
		var submissionContextType string
		var submissionContextID, submissionHomeTeamID, submissionAwayTeamID int64
		if err := tx.QueryRowContext(ctx, verifySubmissionStmt, *input.ResultSubmissionID).Scan(
			&submissionContextType,
			&submissionContextID,
			&submissionHomeTeamID,
			&submissionAwayTeamID,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return hierarchy.ReplayIngestionResult{}, fmt.Errorf("%w: result submission not found", hierarchy.ErrDependency)
			}
			return hierarchy.ReplayIngestionResult{}, err
		}
		if submissionContextType != input.ContextType || submissionContextID != input.ContextID {
			return hierarchy.ReplayIngestionResult{}, fmt.Errorf("%w: result submission context mismatch", hierarchy.ErrConflict)
		}
		if input.SubmittedByTeamID != submissionHomeTeamID && input.SubmittedByTeamID != submissionAwayTeamID {
			return hierarchy.ReplayIngestionResult{}, fmt.Errorf("%w: linkedByTeamId must be submission participant", hierarchy.ErrConflict)
		}

		const linkStmt = `
INSERT INTO result_submission_replay_links(result_submission_id, replay_evidence_id, linked_by_team_id)
VALUES ($1, $2, $3)
ON CONFLICT (result_submission_id, replay_evidence_id) DO NOTHING;`
		if _, err := tx.ExecContext(ctx, linkStmt, *input.ResultSubmissionID, result.Evidence.ID, input.SubmittedByTeamID); err != nil {
			return hierarchy.ReplayIngestionResult{}, mapSQLError(err)
		}
		result.LinkedSubmissionID = input.ResultSubmissionID
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ReplayIngestionResult{}, err
	}
	return result, nil
}

func (s *HierarchyStore) ListReplayEvidence(ctx context.Context) ([]hierarchy.ReplayEvidence, error) {
	const stmt = `
SELECT id, context_type, context_id, submitted_by_team_id, replay_sha256, content_size_bytes, storage_ref, state, created_at
FROM replay_evidence
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evidence := make([]hierarchy.ReplayEvidence, 0)
	for rows.Next() {
		var item hierarchy.ReplayEvidence
		if err := rows.Scan(
			&item.ID,
			&item.ContextType,
			&item.ContextID,
			&item.SubmittedByTeamID,
			&item.ReplaySHA256,
			&item.ContentSizeBytes,
			&item.StorageRef,
			&item.State,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		evidence = append(evidence, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return evidence, nil
}

func (s *HierarchyStore) ListReplayParseRuns(ctx context.Context) ([]hierarchy.ReplayParseRun, error) {
	const stmt = `
SELECT id, replay_evidence_id, parser_name, parser_version, parser_config_digest, status, output_json, created_at
FROM replay_parse_runs
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := make([]hierarchy.ReplayParseRun, 0)
	for rows.Next() {
		var run hierarchy.ReplayParseRun
		if err := rows.Scan(
			&run.ID,
			&run.ReplayEvidenceID,
			&run.ParserName,
			&run.ParserVersion,
			&run.ParserConfigDigest,
			&run.Status,
			&run.OutputJSON,
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

func (s *HierarchyStore) ListResultSubmissionReplayLinks(ctx context.Context) ([]hierarchy.ResultSubmissionReplayLink, error) {
	const stmt = `
SELECT id, result_submission_id, replay_evidence_id, linked_by_team_id, created_at
FROM result_submission_replay_links
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]hierarchy.ResultSubmissionReplayLink, 0)
	for rows.Next() {
		var link hierarchy.ResultSubmissionReplayLink
		if err := rows.Scan(
			&link.ID,
			&link.ResultSubmissionID,
			&link.ReplayEvidenceID,
			&link.LinkedByTeamID,
			&link.CreatedAt,
		); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

func (s *HierarchyStore) ReportException(ctx context.Context, input hierarchy.ReportExceptionInput) (hierarchy.ExceptionTicket, error) {
	if err := hierarchy.ValidateReportExceptionInput(input); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	if err := ensureExceptionContextExists(ctx, s.db, input.ContextType, input.ContextID); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	ticket, err := insertExceptionTicket(ctx, tx, input)
	if err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO exception_actions(ticket_id, action_type, actor, automated, notes, minutes_spent) VALUES ($1, 'reported', $2, FALSE, $3, 0)`,
		ticket.ID,
		"system",
		ticket.ReasonCode,
	); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	return ticket, nil
}

func (s *HierarchyStore) ListOperatorInbox(ctx context.Context) ([]hierarchy.ExceptionTicket, error) {
	const stmt = `
SELECT id, category, context_type, context_id, reported_by_team_id, state, reason_code, severity, suggested_action, details_json, resolution_code, opened_at, triaged_at, resolved_at
FROM exception_tickets
WHERE state <> 'resolved'
ORDER BY severity DESC, opened_at ASC, id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := make([]hierarchy.ExceptionTicket, 0)
	for rows.Next() {
		var ticket hierarchy.ExceptionTicket
		if err := rows.Scan(
			&ticket.ID,
			&ticket.Category,
			&ticket.ContextType,
			&ticket.ContextID,
			&ticket.ReportedByTeamID,
			&ticket.State,
			&ticket.ReasonCode,
			&ticket.Severity,
			&ticket.SuggestedAction,
			&ticket.DetailsJSON,
			&ticket.ResolutionCode,
			&ticket.OpenedAt,
			&ticket.TriagedAt,
			&ticket.ResolvedAt,
		); err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tickets, nil
}

func (s *HierarchyStore) TriageException(ctx context.Context, input hierarchy.TriageExceptionInput) (hierarchy.ExceptionTicket, error) {
	if err := hierarchy.ValidateTriageExceptionInput(input); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const stmt = `
UPDATE exception_tickets
SET
	state = 'triaged',
	reason_code = $2,
	severity = $3,
	suggested_action = $4,
	triaged_at = NOW()
WHERE id = $1
  AND state <> 'resolved'
RETURNING id, category, context_type, context_id, reported_by_team_id, state, reason_code, severity, suggested_action, details_json, resolution_code, opened_at, triaged_at, resolved_at;`
	var ticket hierarchy.ExceptionTicket
	err = tx.QueryRowContext(ctx, stmt, input.TicketID, input.ReasonCode, input.Severity, input.SuggestedAction).Scan(
		&ticket.ID,
		&ticket.Category,
		&ticket.ContextType,
		&ticket.ContextID,
		&ticket.ReportedByTeamID,
		&ticket.State,
		&ticket.ReasonCode,
		&ticket.Severity,
		&ticket.SuggestedAction,
		&ticket.DetailsJSON,
		&ticket.ResolutionCode,
		&ticket.OpenedAt,
		&ticket.TriagedAt,
		&ticket.ResolvedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.ExceptionTicket{}, fmt.Errorf("%w: triage not allowed or ticket not found", hierarchy.ErrConflict)
		}
		return hierarchy.ExceptionTicket{}, mapSQLError(err)
	}

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO exception_actions(ticket_id, action_type, actor, automated, notes, minutes_spent) VALUES ($1, 'triaged', $2, FALSE, $3, $4)`,
		ticket.ID,
		input.Actor,
		input.ReasonCode,
		input.MinutesSpent,
	); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	return ticket, nil
}

func (s *HierarchyStore) ResolveException(ctx context.Context, input hierarchy.ResolveExceptionInput) (hierarchy.ExceptionTicket, error) {
	if err := hierarchy.ValidateResolveExceptionInput(input); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const stmt = `
UPDATE exception_tickets
SET
	state = 'resolved',
	resolution_code = $2,
	resolved_at = NOW()
WHERE id = $1
  AND state <> 'resolved'
RETURNING id, category, context_type, context_id, reported_by_team_id, state, reason_code, severity, suggested_action, details_json, resolution_code, opened_at, triaged_at, resolved_at;`
	var ticket hierarchy.ExceptionTicket
	err = tx.QueryRowContext(ctx, stmt, input.TicketID, input.ResolutionCode).Scan(
		&ticket.ID,
		&ticket.Category,
		&ticket.ContextType,
		&ticket.ContextID,
		&ticket.ReportedByTeamID,
		&ticket.State,
		&ticket.ReasonCode,
		&ticket.Severity,
		&ticket.SuggestedAction,
		&ticket.DetailsJSON,
		&ticket.ResolutionCode,
		&ticket.OpenedAt,
		&ticket.TriagedAt,
		&ticket.ResolvedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return hierarchy.ExceptionTicket{}, fmt.Errorf("%w: resolve not allowed or ticket not found", hierarchy.ErrConflict)
		}
		return hierarchy.ExceptionTicket{}, mapSQLError(err)
	}

	actionType := "resolved"
	if input.Automated {
		actionType = "auto_resolved"
	}
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO exception_actions(ticket_id, action_type, actor, automated, notes, minutes_spent) VALUES ($1, $2, $3, $4, $5, $6)`,
		ticket.ID,
		actionType,
		input.Actor,
		input.Automated,
		input.Notes,
		input.MinutesSpent,
	); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}

	if err := tx.Commit(); err != nil {
		return hierarchy.ExceptionTicket{}, err
	}
	return ticket, nil
}

func (s *HierarchyStore) ListExceptionActions(ctx context.Context) ([]hierarchy.ExceptionAction, error) {
	const stmt = `
SELECT id, ticket_id, action_type, actor, automated, notes, minutes_spent, created_at
FROM exception_actions
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	actions := make([]hierarchy.ExceptionAction, 0)
	for rows.Next() {
		var action hierarchy.ExceptionAction
		if err := rows.Scan(
			&action.ID,
			&action.TicketID,
			&action.ActionType,
			&action.Actor,
			&action.Automated,
			&action.Notes,
			&action.MinutesSpent,
			&action.CreatedAt,
		); err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return actions, nil
}

func (s *HierarchyStore) GetExceptionMetrics(ctx context.Context) (hierarchy.ExceptionMetrics, error) {
	metrics := hierarchy.ExceptionMetrics{}

	const adminHoursStmt = `
SELECT COALESCE(SUM(minutes_spent), 0)::FLOAT8 / 60.0
FROM exception_actions
WHERE automated = FALSE
  AND created_at >= NOW() - INTERVAL '7 days';`
	if err := s.db.QueryRowContext(ctx, adminHoursStmt).Scan(&metrics.AdminHoursPerWeek); err != nil {
		return hierarchy.ExceptionMetrics{}, err
	}

	const touchesStmt = `
WITH total_matches AS (
	SELECT COUNT(*)::FLOAT8 AS c FROM matches
), manual_touches AS (
	SELECT COUNT(a.id)::FLOAT8 AS c
	FROM exception_actions a
	JOIN exception_tickets t ON t.id = a.ticket_id
	WHERE a.automated = FALSE
	  AND t.context_type = 'match'
)
SELECT
	CASE
		WHEN tm.c = 0 THEN 0.0
		ELSE mt.c / tm.c
	END
FROM total_matches tm, manual_touches mt;`
	if err := s.db.QueryRowContext(ctx, touchesStmt).Scan(&metrics.ManualTouchesPerFixture); err != nil {
		return hierarchy.ExceptionMetrics{}, err
	}

	const zeroTouchStmt = `
WITH total_matches AS (
	SELECT COUNT(*)::FLOAT8 AS c FROM matches
), touched_matches AS (
	SELECT COUNT(DISTINCT t.context_id)::FLOAT8 AS c
	FROM exception_actions a
	JOIN exception_tickets t ON t.id = a.ticket_id
	WHERE a.automated = FALSE
	  AND t.context_type = 'match'
)
SELECT
	CASE
		WHEN tm.c = 0 THEN 1.0
		ELSE GREATEST((tm.c - touched.c) / tm.c, 0.0)
	END
FROM total_matches tm, touched_matches touched;`
	if err := s.db.QueryRowContext(ctx, zeroTouchStmt).Scan(&metrics.ZeroTouchFixtureRate); err != nil {
		return hierarchy.ExceptionMetrics{}, err
	}

	const p50Stmt = `
SELECT COALESCE(
	EXTRACT(EPOCH FROM percentile_cont(0.5) WITHIN GROUP (ORDER BY (resolved_at - opened_at))) / 3600.0,
	0.0
)
FROM exception_tickets
WHERE state = 'resolved'
  AND resolved_at IS NOT NULL;`
	if err := s.db.QueryRowContext(ctx, p50Stmt).Scan(&metrics.TimeToCloseHoursP50); err != nil {
		return hierarchy.ExceptionMetrics{}, err
	}

	return metrics, nil
}

func (s *HierarchyStore) EvaluateSchedulingException(ctx context.Context, input hierarchy.EvaluateSchedulingExceptionInput) (hierarchy.ExceptionAutomationResult, error) {
	if err := hierarchy.ValidateEvaluateSchedulingExceptionInput(input); err != nil {
		return hierarchy.ExceptionAutomationResult{}, err
	}

	reportedByTeamID := (*int64)(nil)
	reportInput := hierarchy.ReportExceptionInput{
		Category:         "scheduling_conflict",
		ContextType:      "match",
		ContextID:        input.MatchID,
		ReportedByTeamID: reportedByTeamID,
		ReasonCode:       input.ConflictCode,
		Severity:         3,
		SuggestedAction:  "request_reschedule_confirmation",
		DetailsJSON:      []byte(`{}`),
	}
	ticket, err := s.ReportException(ctx, reportInput)
	if err != nil {
		return hierarchy.ExceptionAutomationResult{}, err
	}

	result := hierarchy.ExceptionAutomationResult{Ticket: ticket}
	if input.HomeConfirmed && input.AwayConfirmed {
		resolved, err := s.ResolveException(ctx, hierarchy.ResolveExceptionInput{
			TicketID:       ticket.ID,
			Actor:          input.Actor,
			ResolutionCode: "reschedule_confirmed",
			Notes:          "both teams confirmed schedule",
			Automated:      true,
			MinutesSpent:   0,
		})
		if err != nil {
			return hierarchy.ExceptionAutomationResult{}, err
		}
		result.Ticket = resolved
		result.AutoResolved = true
	}
	return result, nil
}

func (s *HierarchyStore) EvaluateNoShowException(ctx context.Context, input hierarchy.EvaluateNoShowExceptionInput) (hierarchy.ExceptionAutomationResult, error) {
	if err := hierarchy.ValidateEvaluateNoShowExceptionInput(input); err != nil {
		return hierarchy.ExceptionAutomationResult{}, err
	}

	reasonCode := "attendance_uncertain"
	suggestedAction := "manual_review"
	severity := int32(3)
	autoResolve := false
	resolutionCode := ""
	if input.GraceMinutes >= 15 {
		if input.HomeCheckedIn && !input.AwayCheckedIn {
			reasonCode = "away_no_show"
			suggestedAction = "forfeit_away_team"
			severity = 5
			autoResolve = true
			resolutionCode = "forfeit_away"
		} else if !input.HomeCheckedIn && input.AwayCheckedIn {
			reasonCode = "home_no_show"
			suggestedAction = "forfeit_home_team"
			severity = 5
			autoResolve = true
			resolutionCode = "forfeit_home"
		}
	}

	ticket, err := s.ReportException(ctx, hierarchy.ReportExceptionInput{
		Category:        "no_show",
		ContextType:     "match",
		ContextID:       input.MatchID,
		ReasonCode:      reasonCode,
		Severity:        severity,
		SuggestedAction: suggestedAction,
		DetailsJSON:     []byte(`{}`),
	})
	if err != nil {
		return hierarchy.ExceptionAutomationResult{}, err
	}

	result := hierarchy.ExceptionAutomationResult{Ticket: ticket}
	if autoResolve {
		resolved, err := s.ResolveException(ctx, hierarchy.ResolveExceptionInput{
			TicketID:       ticket.ID,
			Actor:          input.Actor,
			ResolutionCode: resolutionCode,
			Notes:          "rule-based no-show forfeit recommendation applied",
			Automated:      true,
			MinutesSpent:   0,
		})
		if err != nil {
			return hierarchy.ExceptionAutomationResult{}, err
		}
		result.Ticket = resolved
		result.AutoResolved = true
	}
	return result, nil
}

func (s *HierarchyStore) EvaluateReplayDisputeException(ctx context.Context, input hierarchy.EvaluateReplayDisputeExceptionInput) (hierarchy.ExceptionAutomationResult, error) {
	if err := hierarchy.ValidateEvaluateReplayDisputeExceptionInput(input); err != nil {
		return hierarchy.ExceptionAutomationResult{}, err
	}

	category := "result_dispute"
	reasonCode := "dispute_reported"
	suggestedAction := "manual_review"
	severity := int32(4)
	autoResolve := false
	resolutionCode := ""

	if input.ParseStatus == "failed" {
		category = "replay_parse_failure"
		reasonCode = "parse_failed"
		suggestedAction = "reupload_replay"
		severity = 5
	} else if input.IdentityStatus == "mismatch" {
		category = "replay_identity_mismatch"
		reasonCode = "identity_mismatch"
		suggestedAction = "verify_platform_links"
		severity = 5
	} else if !input.DisputeRaised {
		reasonCode = "validated_no_dispute"
		suggestedAction = "no_action_required"
		severity = 1
		autoResolve = true
		resolutionCode = "no_dispute_detected"
	}

	ticket, err := s.ReportException(ctx, hierarchy.ReportExceptionInput{
		Category:        category,
		ContextType:     "result_submission",
		ContextID:       input.ResultSubmissionID,
		ReasonCode:      reasonCode,
		Severity:        severity,
		SuggestedAction: suggestedAction,
		DetailsJSON:     []byte(`{}`),
	})
	if err != nil {
		return hierarchy.ExceptionAutomationResult{}, err
	}

	result := hierarchy.ExceptionAutomationResult{Ticket: ticket}
	if autoResolve {
		resolved, err := s.ResolveException(ctx, hierarchy.ResolveExceptionInput{
			TicketID:       ticket.ID,
			Actor:          input.Actor,
			ResolutionCode: resolutionCode,
			Notes:          "replay and identity checks passed without dispute",
			Automated:      true,
			MinutesSpent:   0,
		})
		if err != nil {
			return hierarchy.ExceptionAutomationResult{}, err
		}
		result.Ticket = resolved
		result.AutoResolved = true
	}
	return result, nil
}

func ensureExceptionContextExists(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, contextType string, contextID int64) error {
	var stmt string
	switch contextType {
	case "match":
		stmt = `SELECT id FROM matches WHERE id = $1;`
	case "scrim":
		stmt = `SELECT id FROM scrims WHERE id = $1;`
	case "result_submission":
		stmt = `SELECT id FROM result_submissions WHERE id = $1;`
	case "replay_evidence":
		stmt = `SELECT id FROM replay_evidence WHERE id = $1;`
	default:
		return fmt.Errorf("%w: unsupported context type", hierarchy.ErrInvalidInput)
	}

	var id int64
	if err := queryer.QueryRowContext(ctx, stmt, contextID).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: exception context not found", hierarchy.ErrDependency)
		}
		return err
	}
	return nil
}

func insertExceptionTicket(ctx context.Context, tx *sql.Tx, input hierarchy.ReportExceptionInput) (hierarchy.ExceptionTicket, error) {
	details := input.DetailsJSON
	if len(details) == 0 {
		details = []byte(`{}`)
	}

	const stmt = `
INSERT INTO exception_tickets(
	category,
	context_type,
	context_id,
	reported_by_team_id,
	reason_code,
	severity,
	suggested_action,
	details_json
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, category, context_type, context_id, reported_by_team_id, state, reason_code, severity, suggested_action, details_json, resolution_code, opened_at, triaged_at, resolved_at;`
	var ticket hierarchy.ExceptionTicket
	err := tx.QueryRowContext(
		ctx,
		stmt,
		input.Category,
		input.ContextType,
		input.ContextID,
		input.ReportedByTeamID,
		input.ReasonCode,
		input.Severity,
		input.SuggestedAction,
		details,
	).Scan(
		&ticket.ID,
		&ticket.Category,
		&ticket.ContextType,
		&ticket.ContextID,
		&ticket.ReportedByTeamID,
		&ticket.State,
		&ticket.ReasonCode,
		&ticket.Severity,
		&ticket.SuggestedAction,
		&ticket.DetailsJSON,
		&ticket.ResolutionCode,
		&ticket.OpenedAt,
		&ticket.TriagedAt,
		&ticket.ResolvedAt,
	)
	if err != nil {
		return hierarchy.ExceptionTicket{}, mapSQLError(err)
	}
	return ticket, nil
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

func resolveContextTeams(ctx context.Context, queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, contextType string, contextID int64) (int64, int64, error) {
	var stmt string
	switch contextType {
	case "scrim":
		stmt = `SELECT home_team_id, away_team_id FROM scrims WHERE id = $1;`
	case "match":
		stmt = `SELECT home_team_id, away_team_id FROM matches WHERE id = $1;`
	default:
		return 0, 0, fmt.Errorf("%w: unsupported context type", hierarchy.ErrInvalidInput)
	}

	var homeTeamID, awayTeamID int64
	if err := queryer.QueryRowContext(ctx, stmt, contextID).Scan(&homeTeamID, &awayTeamID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, fmt.Errorf("%w: context not found", hierarchy.ErrDependency)
		}
		return 0, 0, err
	}
	return homeTeamID, awayTeamID, nil
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

func (s *HierarchyStore) CreateSeason(ctx context.Context, input hierarchy.CreateSeasonInput) (hierarchy.Season, error) {
	if err := hierarchy.ValidateCreateSeasonInput(input); err != nil {
		return hierarchy.Season{}, err
	}

	const stmt = `
INSERT INTO seasons(name, slug)
VALUES ($1, $2)
RETURNING id, name, slug, is_active, created_at;`
	var season hierarchy.Season
	err := s.db.QueryRowContext(ctx, stmt, input.Name, input.Slug).Scan(
		&season.ID,
		&season.Name,
		&season.Slug,
		&season.IsActive,
		&season.CreatedAt,
	)
	if err != nil {
		return hierarchy.Season{}, mapSQLError(err)
	}
	return season, nil
}

func (s *HierarchyStore) ListSeasons(ctx context.Context) ([]hierarchy.Season, error) {
	const stmt = `
SELECT id, name, slug, is_active, created_at
FROM seasons
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seasons := make([]hierarchy.Season, 0)
	for rows.Next() {
		var season hierarchy.Season
		if err := rows.Scan(&season.ID, &season.Name, &season.Slug, &season.IsActive, &season.CreatedAt); err != nil {
			return nil, err
		}
		seasons = append(seasons, season)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return seasons, nil
}

func (s *HierarchyStore) CreateScheduleGroup(ctx context.Context, input hierarchy.CreateScheduleGroupInput) (hierarchy.ScheduleGroup, error) {
	if err := hierarchy.ValidateCreateScheduleGroupInput(input); err != nil {
		return hierarchy.ScheduleGroup{}, err
	}

	const stmt = `
INSERT INTO schedule_groups(season_id, name, sequence)
VALUES ($1, $2, $3)
RETURNING id, season_id, name, sequence, is_active, created_at;`
	var group hierarchy.ScheduleGroup
	err := s.db.QueryRowContext(ctx, stmt, input.SeasonID, input.Name, input.Sequence).Scan(
		&group.ID,
		&group.SeasonID,
		&group.Name,
		&group.Sequence,
		&group.IsActive,
		&group.CreatedAt,
	)
	if err != nil {
		return hierarchy.ScheduleGroup{}, mapSQLError(err)
	}
	return group, nil
}

func (s *HierarchyStore) ListScheduleGroups(ctx context.Context) ([]hierarchy.ScheduleGroup, error) {
	const stmt = `
SELECT id, season_id, name, sequence, is_active, created_at
FROM schedule_groups
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := make([]hierarchy.ScheduleGroup, 0)
	for rows.Next() {
		var group hierarchy.ScheduleGroup
		if err := rows.Scan(&group.ID, &group.SeasonID, &group.Name, &group.Sequence, &group.IsActive, &group.CreatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

func (s *HierarchyStore) CreateFixture(ctx context.Context, input hierarchy.CreateFixtureInput) (hierarchy.Fixture, error) {
	if err := hierarchy.ValidateCreateFixtureInput(input); err != nil {
		return hierarchy.Fixture{}, err
	}

	const stmt = `
INSERT INTO fixtures(schedule_group_id, home_club_id, away_club_id)
VALUES ($1, $2, $3)
RETURNING id, schedule_group_id, home_club_id, away_club_id, is_active, created_at;`
	var fixture hierarchy.Fixture
	err := s.db.QueryRowContext(ctx, stmt, input.ScheduleGroupID, input.HomeClubID, input.AwayClubID).Scan(
		&fixture.ID,
		&fixture.ScheduleGroupID,
		&fixture.HomeClubID,
		&fixture.AwayClubID,
		&fixture.IsActive,
		&fixture.CreatedAt,
	)
	if err != nil {
		return hierarchy.Fixture{}, mapSQLError(err)
	}
	return fixture, nil
}

func (s *HierarchyStore) ListFixtures(ctx context.Context) ([]hierarchy.Fixture, error) {
	const stmt = `
SELECT id, schedule_group_id, home_club_id, away_club_id, is_active, created_at
FROM fixtures
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fixtures := make([]hierarchy.Fixture, 0)
	for rows.Next() {
		var fixture hierarchy.Fixture
		if err := rows.Scan(&fixture.ID, &fixture.ScheduleGroupID, &fixture.HomeClubID, &fixture.AwayClubID, &fixture.IsActive, &fixture.CreatedAt); err != nil {
			return nil, err
		}
		fixtures = append(fixtures, fixture)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return fixtures, nil
}

func (s *HierarchyStore) CreateMatch(ctx context.Context, input hierarchy.CreateMatchInput) (hierarchy.Match, error) {
	if err := hierarchy.ValidateCreateMatchInput(input); err != nil {
		return hierarchy.Match{}, err
	}

	const stmt = `
INSERT INTO matches(
	fixture_id, home_team_id, away_team_id, state, scheduled_for, home_time_ratified_at, away_time_ratified_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, fixture_id, home_team_id, away_team_id, state, scheduled_for, home_time_ratified_at, away_time_ratified_at, created_at;`
	var match hierarchy.Match
	err := s.db.QueryRowContext(
		ctx,
		stmt,
		input.FixtureID,
		input.HomeTeamID,
		input.AwayTeamID,
		input.State,
		input.ScheduledFor,
		input.HomeTimeRatifiedAt,
		input.AwayTimeRatifiedAt,
	).Scan(
		&match.ID,
		&match.FixtureID,
		&match.HomeTeamID,
		&match.AwayTeamID,
		&match.State,
		&match.ScheduledFor,
		&match.HomeTimeRatifiedAt,
		&match.AwayTimeRatifiedAt,
		&match.CreatedAt,
	)
	if err != nil {
		return hierarchy.Match{}, mapSQLError(err)
	}
	return match, nil
}

func (s *HierarchyStore) ListMatches(ctx context.Context) ([]hierarchy.Match, error) {
	const stmt = `
SELECT id, fixture_id, home_team_id, away_team_id, state, scheduled_for, home_time_ratified_at, away_time_ratified_at, created_at
FROM matches
ORDER BY id ASC;`
	rows, err := s.db.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := make([]hierarchy.Match, 0)
	for rows.Next() {
		var match hierarchy.Match
		if err := rows.Scan(
			&match.ID,
			&match.FixtureID,
			&match.HomeTeamID,
			&match.AwayTeamID,
			&match.State,
			&match.ScheduledFor,
			&match.HomeTimeRatifiedAt,
			&match.AwayTimeRatifiedAt,
			&match.CreatedAt,
		); err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return matches, nil
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
