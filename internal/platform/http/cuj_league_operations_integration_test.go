package http

import (
	stdhttp "net/http"
	"testing"
)

func TestCUJLeagueOperations(t *testing.T) {
	app := newCUJApp(t)
	fixture := app.buildLeagueFixture(t, "league-ops")

	app.postAsAdmin(t, "/v1/queue-bans", map[string]any{
		"queueId":  fixture.QueueID,
		"playerId": fixture.HomePlayer,
		"actor":    "league-operator",
		"reason":   "temporary moderation action",
	}, stdhttp.StatusCreated)
	app.post(t, "/v1/queue-entries", map[string]any{
		"queueId": fixture.QueueID,
		"teamId":  fixture.HomeTeamID,
	}, stdhttp.StatusConflict)
	app.postAsAdmin(t, "/v1/queue-bans/lift", map[string]any{
		"queueId":  fixture.QueueID,
		"playerId": fixture.HomePlayer,
		"actor":    "league-operator",
		"reason":   "appeal accepted",
	}, stdhttp.StatusOK)
	app.post(t, "/v1/queue-entries", map[string]any{
		"queueId": fixture.QueueID,
		"teamId":  fixture.HomeTeamID,
	}, stdhttp.StatusCreated)

	season := app.post(t, "/v1/seasons", map[string]any{
		"name": "Ops Season",
		"slug": "ops-season-" + itoa(app.suffix),
	}, stdhttp.StatusCreated)
	group := app.post(t, "/v1/schedule-groups", map[string]any{
		"seasonId": cujID(t, season, "id"),
		"name":     "Ops Week",
		"sequence": 1,
	}, stdhttp.StatusCreated)
	scheduledFixture := app.post(t, "/v1/fixtures", map[string]any{
		"scheduleGroupId": cujID(t, group, "id"),
		"homeClubId":      fixture.HomeClubID,
		"awayClubId":      fixture.AwayClubID,
	}, stdhttp.StatusCreated)
	match := app.post(t, "/v1/matches", map[string]any{
		"fixtureId":  cujID(t, scheduledFixture, "id"),
		"homeTeamId": fixture.HomeTeamID,
		"awayTeamId": fixture.AwayTeamID,
		"state":      "planned",
	}, stdhttp.StatusCreated)
	matchID := cujID(t, match, "id")

	ticket := app.post(t, "/v1/exceptions/report", map[string]any{
		"category":         "scheduling_conflict",
		"contextType":      "match",
		"contextId":        matchID,
		"reportedByTeamId": fixture.HomeTeamID,
		"reasonCode":       "availability_conflict",
		"severity":         3,
		"suggestedAction":  "request_reschedule",
		"detailsJson": map[string]any{
			"source": "CUJ league operations",
		},
	}, stdhttp.StatusCreated)
	ticketID := cujID(t, ticket, "id")
	if cujString(t, ticket, "state") != "open" {
		t.Fatalf("expected reported ticket to be open, got %#v", ticket)
	}

	triaged := app.postAsAdmin(t, "/v1/operator-inbox/triage", map[string]any{
		"ticketId":        ticketID,
		"actor":           "league-operator",
		"reasonCode":      "captains_conflict",
		"severity":        2,
		"suggestedAction": "offer_two_slots",
		"minutesSpent":    5,
	}, stdhttp.StatusOK)
	if cujString(t, triaged, "state") != "triaged" {
		t.Fatalf("expected triaged ticket, got %#v", triaged)
	}

	resolved := app.postAsAdmin(t, "/v1/operator-inbox/resolve", map[string]any{
		"ticketId":       ticketID,
		"actor":          "league-operator",
		"resolutionCode": "rescheduled",
		"notes":          "captains agreed",
		"automated":      false,
		"minutesSpent":   10,
	}, stdhttp.StatusOK)
	if cujString(t, resolved, "state") != "resolved" {
		t.Fatalf("expected resolved ticket, got %#v", resolved)
	}

	if inbox := app.getList(t, "/v1/operator-inbox", stdhttp.StatusOK); len(inbox) == 0 {
		t.Fatal("expected operator inbox to expose league operations tickets")
	}
	if actions := app.getList(t, "/v1/exception-actions", stdhttp.StatusOK); len(actions) < 2 {
		t.Fatalf("expected triage and resolve audit actions, got %d", len(actions))
	}
	metrics := app.getObject(t, "/v1/exception-metrics", stdhttp.StatusOK)
	if _, ok := metrics["manualTouchesPerFixture"]; !ok {
		t.Fatalf("expected exception metrics for operations dashboard, got %#v", metrics)
	}
}
