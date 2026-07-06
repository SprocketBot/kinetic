package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestCUJPlayerQueueJoinLeaveMultipleIdentitiesAndQueues(t *testing.T) {
	app := newCUJApp(t)
	fixture := app.buildLeagueFixture(t, "player-queue")

	secondQueue := app.post(t, "/v1/queues", map[string]any{
		"name": fmt.Sprintf("CUJ Queue secondary %d", app.suffix),
		"slug": fmt.Sprintf("scrim-2v2-player-queue-%d", app.suffix),
	}, stdhttp.StatusCreated)
	secondQueueID := cujID(t, secondQueue, "id")

	homeSubject := fixture.Subject
	awaySubject := "discord:away-player-queue"

	homeQueueOne := postAsSubject(t, app, homeSubject, "/v1/queue-entries", map[string]any{
		"queueId": fixture.QueueID,
		"teamId":  fixture.HomeTeamID,
	}, stdhttp.StatusCreated)
	if cujID(t, homeQueueOne, "queueId") != fixture.QueueID || cujID(t, homeQueueOne, "teamId") != fixture.HomeTeamID {
		t.Fatalf("expected home team queued in primary queue, got %#v", homeQueueOne)
	}

	postAsSubject(t, app, homeSubject, "/v1/queue-entries", map[string]any{
		"queueId": secondQueueID,
		"teamId":  fixture.HomeTeamID,
	}, stdhttp.StatusCreated)

	postAsSubject(t, app, awaySubject, "/v1/queue-entries", map[string]any{
		"queueId": fixture.QueueID,
		"teamId":  fixture.AwayTeamID,
	}, stdhttp.StatusCreated)

	postAsSubject(t, app, homeSubject, "/v1/queue-entries", map[string]any{
		"queueId": fixture.QueueID,
		"teamId":  fixture.HomeTeamID,
	}, stdhttp.StatusConflict)

	deleteAsSubject(t, app, homeSubject, "/v1/queue-entries", map[string]any{
		"queueId": fixture.QueueID,
		"teamId":  fixture.HomeTeamID,
	}, stdhttp.StatusOK)

	postAsSubject(t, app, homeSubject, "/v1/queue-entries", map[string]any{
		"queueId": fixture.QueueID,
		"teamId":  fixture.HomeTeamID,
	}, stdhttp.StatusCreated)

	deleteAsSubject(t, app, homeSubject, "/v1/queue-entries", map[string]any{
		"queueId": secondQueueID,
		"teamId":  fixture.HomeTeamID,
	}, stdhttp.StatusOK)

	deleteAsSubject(t, app, homeSubject, "/v1/queue-entries", map[string]any{
		"queueId": secondQueueID,
		"teamId":  fixture.HomeTeamID,
	}, stdhttp.StatusConflict)

	entries := app.getList(t, "/v1/queue-entries", stdhttp.StatusOK)
	assertHasActiveQueueEntry(t, entries, fixture.QueueID, fixture.HomeTeamID)
	assertHasActiveQueueEntry(t, entries, fixture.QueueID, fixture.AwayTeamID)
	assertNoActiveQueueEntry(t, entries, secondQueueID, fixture.HomeTeamID)
}

func TestCUJPlayerReplaySubmissionRatificationAndRejection(t *testing.T) {
	app := newCUJApp(t)
	fixture := app.buildLeagueFixture(t, "player-replay")

	homeSubject := fixture.Subject
	awaySubject := "discord:away-player-replay"

	ratifiedScrim := app.post(t, "/v1/scrims", map[string]any{
		"queueId":    fixture.QueueID,
		"homeTeamId": fixture.HomeTeamID,
		"awayTeamId": fixture.AwayTeamID,
		"state":      "created",
	}, stdhttp.StatusCreated)
	ratifiedScrimID := cujID(t, ratifiedScrim, "id")

	submission := postAsSubject(t, app, homeSubject, "/v1/result-submissions", map[string]any{
		"contextType":       "scrim",
		"contextId":         ratifiedScrimID,
		"submittedByTeamId": fixture.HomeTeamID,
		"winningTeamId":     fixture.HomeTeamID,
		"losingTeamId":      fixture.AwayTeamID,
		"payloadJson": map[string]any{
			"score": "3-1",
		},
	}, stdhttp.StatusCreated)
	submissionID := cujID(t, submission, "id")
	if got := cujString(t, submission, "submittedBySubject"); got != homeSubject {
		t.Fatalf("expected submittedBySubject %q, got %q", homeSubject, got)
	}

	postAsSubject(t, app, homeSubject, "/v1/result-submissions", map[string]any{
		"contextType":       "scrim",
		"contextId":         ratifiedScrimID,
		"submittedByTeamId": fixture.HomeTeamID,
		"winningTeamId":     fixture.HomeTeamID,
		"losingTeamId":      fixture.AwayTeamID,
		"payloadJson": map[string]any{
			"score": "3-1",
		},
	}, stdhttp.StatusConflict)

	replay := postAsSubject(t, app, homeSubject, "/v1/replay-evidence", replayPayload(ratifiedScrimID, fixture.HomeTeamID, submissionID, "ratify"), stdhttp.StatusCreated)
	if replay["duplicate"].(bool) {
		t.Fatal("expected first replay evidence upload to be non-duplicate")
	}
	if got := cujID(t, replay, "linkedSubmissionId"); got != submissionID {
		t.Fatalf("expected replay evidence linked to submission %d, got %d", submissionID, got)
	}

	postAsSubject(t, app, homeSubject, "/v1/result-submission-ratifications", map[string]any{
		"submissionId": submissionID,
		"teamId":       fixture.AwayTeamID,
	}, stdhttp.StatusConflict)

	ratified := postAsSubject(t, app, awaySubject, "/v1/result-submission-ratifications", map[string]any{
		"submissionId": submissionID,
		"teamId":       fixture.AwayTeamID,
	}, stdhttp.StatusOK)
	if got := cujString(t, ratified, "state"); got != "ratified" {
		t.Fatalf("expected ratified submission state, got %q", got)
	}
	if got := cujString(t, ratified, "awayRatifiedBySubject"); got != awaySubject {
		t.Fatalf("expected away ratification subject %q, got %q", awaySubject, got)
	}

	rejectedScrim := app.post(t, "/v1/scrims", map[string]any{
		"queueId":    fixture.QueueID,
		"homeTeamId": fixture.HomeTeamID,
		"awayTeamId": fixture.AwayTeamID,
		"state":      "created",
	}, stdhttp.StatusCreated)
	rejectedScrimID := cujID(t, rejectedScrim, "id")

	rejectedSubmission := postAsSubject(t, app, homeSubject, "/v1/result-submissions", map[string]any{
		"contextType":       "scrim",
		"contextId":         rejectedScrimID,
		"submittedByTeamId": fixture.HomeTeamID,
		"winningTeamId":     fixture.HomeTeamID,
		"losingTeamId":      fixture.AwayTeamID,
		"payloadJson": map[string]any{
			"score": "2-1",
		},
	}, stdhttp.StatusCreated)
	rejectedSubmissionID := cujID(t, rejectedSubmission, "id")

	postAsSubject(t, app, awaySubject, "/v1/replay-evidence", replayPayload(rejectedScrimID, fixture.AwayTeamID, rejectedSubmissionID, "reject"), stdhttp.StatusCreated)

	rejected := postAsSubject(t, app, awaySubject, "/v1/result-submission-rejections", map[string]any{
		"submissionId": rejectedSubmissionID,
		"teamId":       fixture.AwayTeamID,
		"reason":       "score disputed from replay review",
	}, stdhttp.StatusOK)
	if got := cujString(t, rejected, "state"); got != "rejected" {
		t.Fatalf("expected rejected submission state, got %q", got)
	}

	postAsSubject(t, app, awaySubject, "/v1/result-submission-ratifications", map[string]any{
		"submissionId": rejectedSubmissionID,
		"teamId":       fixture.AwayTeamID,
	}, stdhttp.StatusConflict)
}

func replayPayload(scrimID, teamID, submissionID int64, label string) map[string]any {
	return map[string]any{
		"contextType":        "scrim",
		"contextId":          scrimID,
		"submittedByTeamId":  teamID,
		"replayBody":         "replay-body-" + label,
		"parserName":         "kinetic-rl-parser",
		"parserVersion":      "v0.1.0",
		"parserConfigDigest": "cfg-" + label,
		"resultSubmissionId": submissionID,
		"parseOutputJson": map[string]any{
			"goals": 4,
		},
	}
}

func postAsSubject(t *testing.T, app *cujApp, subject, path string, payload map[string]any, expectedStatus int) map[string]any {
	t.Helper()
	req := requestWithSubject(t, stdhttp.MethodPost, subject, path, payload)
	return app.serveJSON(t, req, path, expectedStatus)
}

func deleteAsSubject(t *testing.T, app *cujApp, subject, path string, payload map[string]any, expectedStatus int) map[string]any {
	t.Helper()
	req := requestWithSubject(t, stdhttp.MethodDelete, subject, path, payload)
	return app.serveJSON(t, req, path, expectedStatus)
}

func requestWithSubject(t *testing.T, method, subject, path string, payload map[string]any) *stdhttp.Request {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal %s payload: %v", path, err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer local:"+subject+":player")
	return req
}

func assertHasActiveQueueEntry(t *testing.T, entries []map[string]any, queueID, teamID int64) {
	t.Helper()
	for _, entry := range entries {
		if cujID(t, entry, "queueId") == queueID && cujID(t, entry, "teamId") == teamID && entry["isActive"] == true {
			return
		}
	}
	t.Fatalf("expected active queue entry for queue=%d team=%d in %#v", queueID, teamID, entries)
}

func assertNoActiveQueueEntry(t *testing.T, entries []map[string]any, queueID, teamID int64) {
	t.Helper()
	for _, entry := range entries {
		if cujID(t, entry, "queueId") == queueID && cujID(t, entry, "teamId") == teamID && entry["isActive"] == true {
			t.Fatalf("did not expect active queue entry for queue=%d team=%d in %#v", queueID, teamID, entries)
		}
	}
}
