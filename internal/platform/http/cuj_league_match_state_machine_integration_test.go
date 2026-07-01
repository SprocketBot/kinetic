package http

import (
	stdhttp "net/http"
	"testing"
)

func TestCUJLeagueMatchStateMachine(t *testing.T) {
	app := newCUJApp(t)
	fixture := app.buildLeagueFixture(t, "league-match")

	season := app.post(t, "/v1/seasons", map[string]any{
		"name": "CUJ Season",
		"slug": "cuj-season-" + itoa(app.suffix),
	}, stdhttp.StatusCreated)
	seasonID := cujID(t, season, "id")

	group := app.post(t, "/v1/schedule-groups", map[string]any{
		"seasonId": seasonID,
		"name":     "Week 1",
		"sequence": 1,
	}, stdhttp.StatusCreated)
	groupID := cujID(t, group, "id")

	scheduledFixture := app.post(t, "/v1/fixtures", map[string]any{
		"scheduleGroupId": groupID,
		"homeClubId":      fixture.HomeClubID,
		"awayClubId":      fixture.AwayClubID,
	}, stdhttp.StatusCreated)
	fixtureID := cujID(t, scheduledFixture, "id")

	app.post(t, "/v1/matches", map[string]any{
		"fixtureId":    fixtureID,
		"homeTeamId":   fixture.HomeTeamID,
		"awayTeamId":   fixture.AwayTeamID,
		"state":        "ready",
		"scheduledFor": "2030-01-01T20:00:00Z",
	}, stdhttp.StatusBadRequest)

	planned := app.post(t, "/v1/matches", map[string]any{
		"fixtureId":  fixtureID,
		"homeTeamId": fixture.HomeTeamID,
		"awayTeamId": fixture.AwayTeamID,
		"state":      "planned",
	}, stdhttp.StatusCreated)
	if cujString(t, planned, "state") != "planned" {
		t.Fatalf("expected planned match, got %#v", planned["state"])
	}

	ready := app.post(t, "/v1/matches", map[string]any{
		"fixtureId":          fixtureID,
		"homeTeamId":         fixture.HomeTeamID,
		"awayTeamId":         fixture.AwayTeamID,
		"state":              "ready",
		"scheduledFor":       "2030-01-01T20:00:00Z",
		"homeTimeRatifiedAt": "2030-01-01T18:00:00Z",
		"awayTimeRatifiedAt": "2030-01-01T19:00:00Z",
	}, stdhttp.StatusCreated)
	matchID := cujID(t, ready, "id")
	if cujString(t, ready, "state") != "ready" {
		t.Fatalf("expected ready match, got %#v", ready["state"])
	}

	submission := app.post(t, "/v1/result-submissions", map[string]any{
		"contextType":       "match",
		"contextId":         matchID,
		"submittedByTeamId": fixture.HomeTeamID,
		"winningTeamId":     fixture.HomeTeamID,
		"losingTeamId":      fixture.AwayTeamID,
		"payloadJson": map[string]any{
			"series": "3-2",
		},
	}, stdhttp.StatusCreated)
	submissionID := cujID(t, submission, "id")

	app.post(t, "/v1/result-submission-ratifications", map[string]any{
		"submissionId": submissionID,
		"teamId":       fixture.HomeTeamID,
	}, stdhttp.StatusOK)
	ratified := app.post(t, "/v1/result-submission-ratifications", map[string]any{
		"submissionId": submissionID,
		"teamId":       fixture.AwayTeamID,
	}, stdhttp.StatusOK)
	if cujString(t, ratified, "state") != "ratified" {
		t.Fatalf("expected match result to ratify after both teams, got %#v", ratified["state"])
	}

	app.post(t, "/v1/result-overrides", map[string]any{
		"submissionId":  submissionID,
		"actor":         "league-admin",
		"reason":        "official review requires correction",
		"winningTeamId": fixture.AwayTeamID,
		"losingTeamId":  fixture.HomeTeamID,
	}, stdhttp.StatusUnauthorized)
	overridden := app.postAsAdmin(t, "/v1/result-overrides", map[string]any{
		"submissionId":  submissionID,
		"actor":         "league-admin",
		"reason":        "official review requires correction",
		"winningTeamId": fixture.AwayTeamID,
		"losingTeamId":  fixture.HomeTeamID,
	}, stdhttp.StatusOK)
	if cujID(t, overridden, "winningTeamId") != fixture.AwayTeamID {
		t.Fatalf("expected override to update winning team, got %#v", overridden)
	}

	if seasons := app.getList(t, "/v1/seasons", stdhttp.StatusOK); len(seasons) == 0 {
		t.Fatal("expected season list for league match page")
	}
	if groups := app.getList(t, "/v1/schedule-groups", stdhttp.StatusOK); len(groups) == 0 {
		t.Fatal("expected schedule group list for league match page")
	}
	if matches := app.getList(t, "/v1/matches", stdhttp.StatusOK); len(matches) < 2 {
		t.Fatalf("expected planned and ready matches in match list, got %d", len(matches))
	}
}
