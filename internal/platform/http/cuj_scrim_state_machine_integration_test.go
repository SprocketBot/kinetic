package http

import (
	stdhttp "net/http"
	"strconv"
	"testing"
)

func TestCUJScrimStateMachine(t *testing.T) {
	app := newCUJApp(t)
	fixture := app.buildLeagueFixture(t, "scrim-state")

	app.post(t, "/v1/queue-entries", map[string]any{
		"queueId": fixture.QueueID,
		"teamId":  fixture.HomeTeamID,
	}, stdhttp.StatusCreated)
	app.post(t, "/v1/queue-entries", map[string]any{
		"queueId": fixture.QueueID,
		"teamId":  fixture.HomeTeamID,
	}, stdhttp.StatusConflict)
	app.post(t, "/v1/queue-entries", map[string]any{
		"queueId": fixture.QueueID,
		"teamId":  fixture.AwayTeamID,
	}, stdhttp.StatusCreated)

	popped := app.postAsAdmin(t, "/v1/scrim-promotions", map[string]any{
		"queueId": fixture.QueueID,
	}, stdhttp.StatusCreated)
	scrimID := cujID(t, popped, "id")
	if cujString(t, popped, "state") != "popped" {
		t.Fatalf("expected promoted scrim to be popped, got %#v", popped["state"])
	}
	if popped["lobbyName"] == nil || popped["lobbyPassword"] == nil {
		t.Fatalf("expected popped scrim to include lobby info, got %#v", popped)
	}

	app.post(t, "/v1/scrim-promotions", map[string]any{
		"queueId": fixture.QueueID,
	}, stdhttp.StatusUnauthorized)

	homeCheckIn := app.post(t, "/v1/scrims/"+itoa(scrimID)+"/check-in", map[string]any{
		"teamId": fixture.HomeTeamID,
	}, stdhttp.StatusOK)
	if cujString(t, homeCheckIn, "state") != "popped" {
		t.Fatalf("expected first check-in to keep scrim popped, got %#v", homeCheckIn["state"])
	}

	awayCheckIn := app.post(t, "/v1/scrims/"+itoa(scrimID)+"/check-in", map[string]any{
		"teamId": fixture.AwayTeamID,
	}, stdhttp.StatusOK)
	if cujString(t, awayCheckIn, "state") != "in_progress" {
		t.Fatalf("expected second check-in to start scrim, got %#v", awayCheckIn["state"])
	}

	submission := app.post(t, "/v1/result-submissions", map[string]any{
		"contextType":       "scrim",
		"contextId":         scrimID,
		"submittedByTeamId": fixture.HomeTeamID,
		"winningTeamId":     fixture.HomeTeamID,
		"losingTeamId":      fixture.AwayTeamID,
		"payloadJson": map[string]any{
			"score": "3-1",
		},
	}, stdhttp.StatusCreated)
	submissionID := cujID(t, submission, "id")
	if cujString(t, submission, "state") != "pending_ratification" {
		t.Fatalf("expected submitted result to require ratification, got %#v", submission["state"])
	}

	replay := app.post(t, "/v1/replay-evidence", map[string]any{
		"contextType":        "scrim",
		"contextId":          scrimID,
		"submittedByTeamId":  fixture.HomeTeamID,
		"replayBody":         "scrim-state-replay-body",
		"parserName":         "sprocket-rl-parser",
		"parserVersion":      "v1",
		"parserConfigDigest": "default",
		"resultSubmissionId": submissionID,
		"parseOutputJson": map[string]any{
			"goals": 4,
		},
	}, stdhttp.StatusCreated)
	if replay["duplicate"].(bool) {
		t.Fatalf("expected first replay upload to be non-duplicate, got %#v", replay)
	}
	if cujID(t, replay, "linkedSubmissionId") != submissionID {
		t.Fatalf("expected replay to link submission %d, got %#v", submissionID, replay["linkedSubmissionId"])
	}

	duplicate := app.post(t, "/v1/replay-evidence", map[string]any{
		"contextType":        "scrim",
		"contextId":          scrimID,
		"submittedByTeamId":  fixture.HomeTeamID,
		"replayBody":         "scrim-state-replay-body",
		"parserName":         "sprocket-rl-parser",
		"parserVersion":      "v1",
		"parserConfigDigest": "default",
		"resultSubmissionId": submissionID,
		"parseOutputJson": map[string]any{
			"goals": 4,
		},
	}, stdhttp.StatusOK)
	if !duplicate["duplicate"].(bool) {
		t.Fatalf("expected second replay upload to be duplicate, got %#v", duplicate)
	}

	app.post(t, "/v1/result-submission-ratifications", map[string]any{
		"submissionId": submissionID,
		"teamId":       fixture.HomeTeamID,
	}, stdhttp.StatusOK)
	ratified := app.post(t, "/v1/result-submission-ratifications", map[string]any{
		"submissionId": submissionID,
		"teamId":       fixture.AwayTeamID,
	}, stdhttp.StatusOK)
	if cujString(t, ratified, "state") != "ratified" {
		t.Fatalf("expected both teams ratifying to ratify result, got %#v", ratified["state"])
	}
}

func itoa(value int64) string {
	return strconv.FormatInt(value, 10)
}
