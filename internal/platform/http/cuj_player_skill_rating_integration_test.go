package http

import (
	stdhttp "net/http"
	"testing"
)

func TestCUJPlayerSkillRating(t *testing.T) {
	app := newCUJApp(t)
	fixture := app.buildLeagueFixture(t, "player-rating")

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
			"score": "3-0",
		},
	}, stdhttp.StatusCreated)
	submissionID := cujID(t, submission, "id")

	app.post(t, "/v1/result-submission-ratifications", map[string]any{
		"submissionId": submissionID,
		"teamId":       fixture.HomeTeamID,
	}, stdhttp.StatusOK)
	app.post(t, "/v1/result-submission-ratifications", map[string]any{
		"submissionId": submissionID,
		"teamId":       fixture.AwayTeamID,
	}, stdhttp.StatusOK)

	ratings := app.getList(t, "/v1/player-ratings", stdhttp.StatusOK)
	if len(ratings) < 2 {
		t.Fatalf("expected ratified scrim to create/update ratings for both rostered players, got %d", len(ratings))
	}
	adjustments := app.getList(t, "/v1/rating-adjustments", stdhttp.StatusOK)
	if len(adjustments) < 2 {
		t.Fatalf("expected ratified scrim to record rating adjustments, got %d", len(adjustments))
	}

	app.post(t, "/v1/player-ratings/adjust", map[string]any{
		"actorPlayerId":  fixture.HomePlayer,
		"targetPlayerId": fixture.HomePlayer,
		"contextKey":     "manual-review",
		"rating":         1100,
		"uncertainty":    250,
		"matchesPlayed":  10,
		"reason":         "self edit must fail",
	}, stdhttp.StatusConflict)

	manual := app.post(t, "/v1/player-ratings/adjust", map[string]any{
		"actorPlayerId":  fixture.HomePlayer,
		"targetPlayerId": fixture.AwayPlayer,
		"contextKey":     "manual-review",
		"rating":         1125,
		"uncertainty":    225,
		"matchesPlayed":  12,
		"reason":         "league review correction",
	}, stdhttp.StatusOK)
	if cujID(t, manual, "playerId") != fixture.AwayPlayer {
		t.Fatalf("expected manual adjustment to update away player, got %#v", manual)
	}
}
