package replaystats

import "time"

type Round struct {
	ID           int64     `json:"id"`
	SubmissionID int64     `json:"submissionId"`
	RoundNumber  int32     `json:"roundNumber"`
	Duration     int32     `json:"duration"` // seconds
	CreatedAt    time.Time `json:"createdAt"`
}

type PlayerStatLine struct {
	ID             int64  `json:"id"`
	RoundID        int64  `json:"roundId"`
	PlayerID       *int64 `json:"playerId,omitempty"`
	ReplayIdentity string `json:"replayIdentity"`
	Goals          int32  `json:"goals"`
	Assists        int32  `json:"assists"`
	Saves          int32  `json:"saves"`
	Shots          int32  `json:"shots"`
	Score          int32  `json:"score"`
	// Computed performance indices (Theme 5E)
	OPI float64 `json:"opi"` // (goals + 0.75*assists + 0.3*shots) / rounds_played
	DPI float64 `json:"dpi"` // saves / rounds_played
	GPI float64 `json:"gpi"` // opi + dpi
}

type TeamStatLine struct {
	ID      int64  `json:"id"`
	RoundID int64  `json:"roundId"`
	TeamID  *int64 `json:"teamId,omitempty"`
	Goals   int32  `json:"goals"`
	Shots   int32  `json:"shots"`
	Saves   int32  `json:"saves"`
}

type ParseReplayInput struct {
	EvidenceID  int64
	ContextType string // "scrim" or "match"
	ContextID   int64
}

type ParseReplayResult struct {
	ParseRunID  int64
	RoundsCount int32
}

type PlayerCareerStats struct {
	PlayerID     int64   `json:"playerId"`
	TotalGames   int32   `json:"totalGames"`
	TotalGoals   int32   `json:"totalGoals"`
	TotalAssists int32   `json:"totalAssists"`
	TotalSaves   int32   `json:"totalSaves"`
	TotalShots   int32   `json:"totalShots"`
	AvgOPI       float64 `json:"avgOpi"`
	AvgDPI       float64 `json:"avgDpi"`
	AvgGPI       float64 `json:"avgGpi"`
}
