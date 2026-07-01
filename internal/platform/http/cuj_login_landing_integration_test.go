package http

import (
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCUJLoginAndLandingPageData(t *testing.T) {
	app := newCUJApp(t)
	fixture := app.buildLeagueFixture(t, "login-landing")

	unauthenticated := httptest.NewRequest(stdhttp.MethodGet, "/v1/session", nil)
	unauthenticatedRR := httptest.NewRecorder()
	app.server.Handler().ServeHTTP(unauthenticatedRR, unauthenticated)
	if unauthenticatedRR.Code != stdhttp.StatusUnauthorized {
		t.Fatalf("expected anonymous session lookup to return 401, got %d", unauthenticatedRR.Code)
	}

	loginReq := httptest.NewRequest(stdhttp.MethodGet, "/v1/auth/login?subject="+fixture.Subject+"&displayName=CUJ%20Player&roles=player&redirect=http://localhost:5173/app/player", nil)
	loginRR := httptest.NewRecorder()
	app.server.Handler().ServeHTTP(loginRR, loginReq)
	if loginRR.Code != stdhttp.StatusFound {
		t.Fatalf("expected local login redirect, got %d body=%s", loginRR.Code, loginRR.Body.String())
	}

	callbackLocation := loginRR.Result().Header.Get("Location")
	if !strings.HasPrefix(callbackLocation, "/v1/auth/callback?") {
		t.Fatalf("expected login redirect to local callback, got %q", callbackLocation)
	}

	callbackReq := httptest.NewRequest(stdhttp.MethodGet, callbackLocation, nil)
	callbackRR := httptest.NewRecorder()
	app.server.Handler().ServeHTTP(callbackRR, callbackReq)
	if callbackRR.Code != stdhttp.StatusFound {
		t.Fatalf("expected callback redirect to web app, got %d body=%s", callbackRR.Code, callbackRR.Body.String())
	}
	if callbackRR.Result().Header.Get("Location") != "http://localhost:5173/app/player" {
		t.Fatalf("expected callback to preserve player landing redirect, got %q", callbackRR.Result().Header.Get("Location"))
	}
	sessionCookies := callbackRR.Result().Cookies()
	if len(sessionCookies) == 0 {
		t.Fatal("expected callback to issue a session cookie")
	}

	sessionReq := httptest.NewRequest(stdhttp.MethodGet, "/v1/session", nil)
	sessionReq.AddCookie(sessionCookies[0])
	sessionRR := httptest.NewRecorder()
	app.server.Handler().ServeHTTP(sessionRR, sessionReq)
	if sessionRR.Code != stdhttp.StatusOK {
		t.Fatalf("expected authenticated session lookup, got %d body=%s", sessionRR.Code, sessionRR.Body.String())
	}
	var session map[string]any
	if err := json.NewDecoder(sessionRR.Body).Decode(&session); err != nil {
		t.Fatalf("failed to decode session response: %v", err)
	}
	if session["subject"] != fixture.Subject {
		t.Fatalf("expected session subject %q, got %#v", fixture.Subject, session["subject"])
	}

	app.post(t, "/v1/queue-entries", map[string]any{
		"queueId": fixture.QueueID,
		"teamId":  fixture.HomeTeamID,
	}, stdhttp.StatusCreated)
	app.post(t, "/v1/scrims", map[string]any{
		"queueId":    fixture.QueueID,
		"homeTeamId": fixture.HomeTeamID,
		"awayTeamId": fixture.AwayTeamID,
		"state":      "created",
	}, stdhttp.StatusCreated)

	if entries := app.getList(t, "/v1/queue-entries", stdhttp.StatusOK); len(entries) == 0 {
		t.Fatal("expected landing page queue status data")
	}
	if scrims := app.getList(t, "/v1/scrims", stdhttp.StatusOK); len(scrims) == 0 {
		t.Fatal("expected landing page scrim status data")
	}
	eligibility := app.getObject(t, "/v1/eligibility?subject="+fixture.Subject, stdhttp.StatusOK)
	if _, ok := eligibility["points"]; !ok {
		t.Fatalf("expected eligibility points for landing page, got %#v", eligibility)
	}
	if _, ok := eligibility["eligibleUntil"]; !ok {
		t.Fatalf("expected eligibility expiration for landing page, got %#v", eligibility)
	}
}
