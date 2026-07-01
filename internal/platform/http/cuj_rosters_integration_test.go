package http

import (
	stdhttp "net/http"
	"testing"
)

func TestCUJFranchiseClubTeamRosters(t *testing.T) {
	app := newCUJApp(t)
	fixture := app.buildLeagueFixture(t, "rosters")

	if leagues := app.getList(t, "/v1/leagues", stdhttp.StatusOK); len(leagues) == 0 {
		t.Fatal("expected public league hierarchy to be visible")
	}
	if clubs := app.getList(t, "/v1/clubs", stdhttp.StatusOK); len(clubs) < 2 {
		t.Fatalf("expected home and away clubs to be visible, got %d", len(clubs))
	}
	if teams := app.getList(t, "/v1/teams", stdhttp.StatusOK); len(teams) < 2 {
		t.Fatalf("expected home and away teams to be visible, got %d", len(teams))
	}
	if memberships := app.getList(t, "/v1/roster-memberships", stdhttp.StatusOK); len(memberships) < 2 {
		t.Fatalf("expected initial active rosters to be visible, got %d", len(memberships))
	}

	app.post(t, "/v1/roster-memberships", map[string]any{
		"playerId": fixture.HomePlayer,
		"teamId":   fixture.HomeTeamID,
	}, stdhttp.StatusConflict)
	app.post(t, "/v1/roster-memberships", map[string]any{
		"playerId": fixture.HomePlayer,
		"teamId":   int64(999999999),
	}, stdhttp.StatusConflict)

	app.postAsAdmin(t, "/v1/role-assignments", map[string]any{
		"actorPlayerId":  fixture.HomePlayer,
		"targetPlayerId": fixture.AwayPlayer,
		"role":           "captain",
		"reason":         "missing scope must fail",
	}, stdhttp.StatusBadRequest)

	role := app.postAsAdmin(t, "/v1/role-assignments", map[string]any{
		"actorPlayerId":  fixture.HomePlayer,
		"targetPlayerId": fixture.AwayPlayer,
		"role":           "captain",
		"teamId":         fixture.AwayTeamID,
		"reason":         "season roster leadership",
	}, stdhttp.StatusCreated)
	roleID := cujID(t, role, "id")
	if cujString(t, role, "role") != "captain" {
		t.Fatalf("expected captain role assignment, got %#v", role)
	}

	app.postAsAdmin(t, "/v1/role-assignments", map[string]any{
		"actorPlayerId":  fixture.HomePlayer,
		"targetPlayerId": fixture.AwayPlayer,
		"role":           "captain",
		"teamId":         fixture.AwayTeamID,
		"reason":         "duplicate active role should fail",
	}, stdhttp.StatusConflict)

	revoked := app.postAsAdmin(t, "/v1/role-assignments/revoke", map[string]any{
		"actorPlayerId": fixture.HomePlayer,
		"assignmentId":  roleID,
		"reason":        "role realignment",
	}, stdhttp.StatusOK)
	if revoked["isActive"].(bool) {
		t.Fatalf("expected revoked role assignment to be inactive, got %#v", revoked)
	}

	if roles := app.getList(t, "/v1/role-assignments", stdhttp.StatusOK); len(roles) == 0 {
		t.Fatal("expected role assignment history to remain visible")
	}
}
