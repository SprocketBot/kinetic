package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sprocketbot/sprocket-v3/internal/domain/replaystats"
)

// ReplayStatsStore implements replaystats.Store backed by a PostgreSQL database.
type ReplayStatsStore struct {
	db *sql.DB
}

// NewReplayStatsStore creates a new ReplayStatsStore.
func NewReplayStatsStore(db *sql.DB) *ReplayStatsStore {
	return &ReplayStatsStore{db: db}
}

// ListStatsBySubmission returns all player stat lines for rounds belonging to
// the given result submission.
func (s *ReplayStatsStore) ListStatsBySubmission(ctx context.Context, submissionID int64) ([]replaystats.PlayerStatLine, error) {
	const stmt = `
SELECT
    psl.id,
    psl.round_id,
    psl.player_id,
    psl.replay_identity,
    psl.goals,
    psl.assists,
    psl.saves,
    psl.shots,
    psl.score,
    COUNT(*) OVER (PARTITION BY psl.player_id) AS rounds_played
FROM player_stat_lines psl
JOIN rounds r ON r.id = psl.round_id
WHERE r.submission_id = $1
ORDER BY psl.id ASC;`

	rows, err := s.db.QueryContext(ctx, stmt, submissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]replaystats.PlayerStatLine, 0)
	for rows.Next() {
		var line replaystats.PlayerStatLine
		var roundsPlayed int64
		if err := rows.Scan(
			&line.ID,
			&line.RoundID,
			&line.PlayerID,
			&line.ReplayIdentity,
			&line.Goals,
			&line.Assists,
			&line.Saves,
			&line.Shots,
			&line.Score,
			&roundsPlayed,
		); err != nil {
			return nil, err
		}
		if roundsPlayed > 0 {
			line.OPI = (float64(line.Goals) + 0.75*float64(line.Assists) + 0.3*float64(line.Shots)) / float64(roundsPlayed)
			line.DPI = float64(line.Saves) / float64(roundsPlayed)
			line.GPI = line.OPI + line.DPI
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// GetPlayerCareerStats returns aggregated career statistics for a player.
func (s *ReplayStatsStore) GetPlayerCareerStats(ctx context.Context, playerID int64) (replaystats.PlayerCareerStats, error) {
	const stmt = `
SELECT
    COUNT(*)           AS total_games,
    COALESCE(SUM(goals), 0)   AS total_goals,
    COALESCE(SUM(assists), 0) AS total_assists,
    COALESCE(SUM(saves), 0)   AS total_saves,
    COALESCE(SUM(shots), 0)   AS total_shots
FROM player_stat_lines
WHERE player_id = $1;`

	var stats replaystats.PlayerCareerStats
	stats.PlayerID = playerID

	var totalGames int64
	var totalGoals, totalAssists, totalSaves, totalShots int64
	err := s.db.QueryRowContext(ctx, stmt, playerID).Scan(
		&totalGames,
		&totalGoals,
		&totalAssists,
		&totalSaves,
		&totalShots,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return replaystats.PlayerCareerStats{PlayerID: playerID}, replaystats.ErrNotFound
		}
		return replaystats.PlayerCareerStats{}, err
	}

	if totalGames == 0 {
		return replaystats.PlayerCareerStats{PlayerID: playerID}, nil
	}

	stats.TotalGames = int32(totalGames)
	stats.TotalGoals = int32(totalGoals)
	stats.TotalAssists = int32(totalAssists)
	stats.TotalSaves = int32(totalSaves)
	stats.TotalShots = int32(totalShots)

	// Compute per-game averages using the SprocketRating formula
	g := float64(totalGames)
	stats.AvgOPI = (float64(totalGoals) + 0.75*float64(totalAssists) + 0.3*float64(totalShots)) / g
	stats.AvgDPI = float64(totalSaves) / g
	stats.AvgGPI = stats.AvgOPI + stats.AvgDPI

	return stats, nil
}
