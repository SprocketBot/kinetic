package http

import (
	"context"
	stdhttp "net/http"
	"testing"
	"time"
)

func TestCUJPerGameStatisticsAndHistory(t *testing.T) {
	app := newCUJApp(t)
	fixture := app.buildLeagueFixture(t, "stats-history")

	scrim := app.post(t, "/v1/scrims", map[string]any{
		"queueId":    fixture.QueueID,
		"homeTeamId": fixture.HomeTeamID,
		"awayTeamId": fixture.AwayTeamID,
		"state":      "created",
	}, stdhttp.StatusCreated)
	scrimID := cujID(t, scrim, "id")

	submission := app.post(t, "/v1/result-submissions", map[string]any{
		"contextType":       "scrim",
		"contextId":         scrimID,
		"submittedByTeamId": fixture.HomeTeamID,
		"winningTeamId":     fixture.HomeTeamID,
		"losingTeamId":      fixture.AwayTeamID,
		"payloadJson": map[string]any{
			"score": "4-2",
		},
	}, stdhttp.StatusCreated)
	submissionID := cujID(t, submission, "id")

	app.post(t, "/v1/replay-evidence", map[string]any{
		"contextType":        "scrim",
		"contextId":          scrimID,
		"submittedByTeamId":  fixture.HomeTeamID,
		"replayBody":         "stats-history-replay-body",
		"parserName":         "sprocket-rl-parser",
		"parserVersion":      "v1",
		"parserConfigDigest": "default",
		"resultSubmissionId": submissionID,
		"parseOutputJson": map[string]any{
			"goals": 4,
		},
	}, stdhttp.StatusCreated)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var roundID int64
	if err := app.conn.QueryRowContext(
		ctx,
		`INSERT INTO rounds(submission_id, round_number, duration_seconds)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		submissionID,
		99,
		300,
	).Scan(&roundID); err != nil {
		t.Fatalf("failed to create deterministic round fixture: %v", err)
	}
	if _, err := app.conn.ExecContext(
		ctx,
		`INSERT INTO player_stat_lines(round_id, player_id, replay_identity, goals, assists, saves, shots, score)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		roundID,
		fixture.HomePlayer,
		"home-player-replay-identity",
		2,
		1,
		3,
		5,
		650,
	); err != nil {
		t.Fatalf("failed to create deterministic player stat fixture: %v", err)
	}

	if evidence := app.getList(t, "/v1/replay-evidence", stdhttp.StatusOK); len(evidence) == 0 {
		t.Fatal("expected replay evidence history to be visible")
	}
	if links := app.getList(t, "/v1/result-submission-replay-links", stdhttp.StatusOK); len(links) == 0 {
		t.Fatal("expected replay-to-submission history links to be visible")
	}
	if stats := app.getList(t, "/v1/result-submissions/"+itoa(submissionID)+"/stats", stdhttp.StatusOK); len(stats) == 0 {
		t.Fatal("expected per-submission stat lines to be visible")
	}

	career := app.getObject(t, "/v1/players/"+itoa(fixture.HomePlayer)+"/career-stats", stdhttp.StatusOK)
	if cujID(t, career, "playerId") != fixture.HomePlayer {
		t.Fatalf("expected career stats for player %d, got %#v", fixture.HomePlayer, career)
	}
	if career["totalGoals"].(float64) < 2 {
		t.Fatalf("expected career goal aggregate from stat fixture, got %#v", career)
	}
}
