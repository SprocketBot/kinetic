package replayparser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Parser turns raw Rocket League replay bytes into the summary emitted by the
// Sprocket rl-replay command.
type Parser interface {
	Parse(ctx context.Context, replay []byte) (Result, error)
}

type Command struct {
	path string
}

type Result struct {
	Summary MatchSummary
	RawJSON json.RawMessage
}

type MatchSummary struct {
	Date            *string  `json:"date"`
	Version         string   `json:"version"`
	Map             *string  `json:"map"`
	GameType        string   `json:"game_type"`
	DurationSeconds *float64 `json:"duration_seconds"`
	Winner          *string  `json:"winner"`
	Teams           []Team   `json:"teams"`
	Players         []Player `json:"players"`
	Goals           []Goal   `json:"goals"`
}

type Team struct {
	Index int32  `json:"index"`
	Name  string `json:"name"`
	Score int32  `json:"score"`
}

type Player struct {
	Identity PlayerIdentity `json:"identity"`
	Stats    PlayerStats    `json:"stats"`
	Team     *int32         `json:"team"`
}

type PlayerIdentity struct {
	Name     string  `json:"name"`
	Platform *string `json:"platform"`
	OnlineID *string `json:"online_id"`
	EpicID   *string `json:"epic_id"`
}

type PlayerStats struct {
	Goals              int32    `json:"goals"`
	Assists            int32    `json:"assists"`
	Shots              int32    `json:"shots"`
	Saves              int32    `json:"saves"`
	Score              int32    `json:"score"`
	ShotsOnTarget      *int32   `json:"shots_on_target"`
	GoalsAgainst       *int32   `json:"goals_against"`
	ShotsAgainst       *int32   `json:"shots_against"`
	ShootingPercentage *float64 `json:"shooting_percentage"`
}

type Goal struct {
	Frame      int32  `json:"frame"`
	PlayerName string `json:"player_name"`
	Team       string `json:"team"`
	GoalNumber int32  `json:"goal_number"`
}

func NewCommand(path string) *Command {
	return &Command{path: strings.TrimSpace(path)}
}

func (c *Command) Parse(ctx context.Context, replay []byte) (Result, error) {
	if c == nil || c.path == "" {
		return Result{}, fmt.Errorf("replay parser path is not configured")
	}
	if len(replay) == 0 {
		return Result{}, fmt.Errorf("replay is empty")
	}

	file, err := os.CreateTemp("", "kinetic-*.replay")
	if err != nil {
		return Result{}, fmt.Errorf("create replay temp file: %w", err)
	}
	path := file.Name()
	defer func() { _ = os.Remove(path) }()

	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return Result{}, fmt.Errorf("secure replay temp file: %w", err)
	}
	if _, err := file.Write(replay); err != nil {
		_ = file.Close()
		return Result{}, fmt.Errorf("write replay temp file: %w", err)
	}
	if err := file.Close(); err != nil {
		return Result{}, fmt.Errorf("close replay temp file: %w", err)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, c.path, path)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return Result{}, fmt.Errorf("rl-replay failed: %s", message)
	}

	raw := bytes.TrimSpace(stdout.Bytes())
	var summary MatchSummary
	if err := json.Unmarshal(raw, &summary); err != nil {
		return Result{}, fmt.Errorf("decode rl-replay output: %w", err)
	}
	if err := validateSummary(summary); err != nil {
		return Result{}, err
	}

	return Result{Summary: summary, RawJSON: append(json.RawMessage(nil), raw...)}, nil
}

func validateSummary(summary MatchSummary) error {
	if strings.TrimSpace(summary.Version) == "" {
		return fmt.Errorf("rl-replay output is missing version")
	}
	if len(summary.Teams) < 2 {
		return fmt.Errorf("rl-replay output must contain both teams")
	}
	seenTeams := map[int32]bool{}
	for _, team := range summary.Teams {
		if team.Index != 0 && team.Index != 1 {
			return fmt.Errorf("rl-replay output contains unsupported team index %d", team.Index)
		}
		seenTeams[team.Index] = true
	}
	if !seenTeams[0] || !seenTeams[1] {
		return fmt.Errorf("rl-replay output must contain team indexes 0 and 1")
	}
	for _, player := range summary.Players {
		if strings.TrimSpace(player.Identity.Name) == "" {
			return fmt.Errorf("rl-replay output contains a player without a name")
		}
		if player.Team == nil || (*player.Team != 0 && *player.Team != 1) {
			return fmt.Errorf("rl-replay output contains an unassigned player %q", player.Identity.Name)
		}
	}
	return nil
}
