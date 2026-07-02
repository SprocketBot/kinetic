package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/kineticbot/kinetic-v3/internal/platform/config"
	"github.com/kineticbot/kinetic-v3/internal/platform/db"
)

type cujApp struct {
	server *Server
	conn   *sql.DB
	suffix int64
}

type cujLeagueFixture struct {
	LeagueID    int64
	FranchiseID int64
	HomeClubID  int64
	AwayClubID  int64
	HomeTeamID  int64
	AwayTeamID  int64
	HomePlayer  int64
	AwayPlayer  int64
	QueueID     int64
	Subject     string
}

func newCUJApp(t *testing.T) *cujApp {
	t.Helper()

	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping CUJ integration test")
	}

	conn, err := db.Open(testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	t.Cleanup(func() {
		_ = conn.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := db.Ping(ctx, conn); err != nil {
		t.Fatalf("failed to ping test DB: %v", err)
	}

	migrator := db.NewMigrator(conn, "../../../migrations")
	if _, err := migrator.Up(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	stores := db.NewStores(conn)
	server := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore:   stores.HierarchyStore,
		ReplayStatsStore: db.NewReplayStatsStore(conn),
	})

	return &cujApp{
		server: server,
		conn:   conn,
		suffix: time.Now().UnixNano(),
	}
}

func (a *cujApp) buildLeagueFixture(t *testing.T, label string) cujLeagueFixture {
	t.Helper()

	slug := fmt.Sprintf("%s-%d", label, a.suffix)
	subject := "discord:" + slug

	league := a.post(t, "/v1/leagues", map[string]any{
		"name": "CUJ League " + label,
		"slug": "league-" + slug,
	}, stdhttp.StatusCreated)
	leagueID := cujID(t, league, "id")

	franchise := a.post(t, "/v1/franchises", map[string]any{
		"leagueId": leagueID,
		"name":     "CUJ Franchise " + label,
		"slug":     "franchise-" + slug,
	}, stdhttp.StatusCreated)
	franchiseID := cujID(t, franchise, "id")

	homeClub := a.post(t, "/v1/clubs", map[string]any{
		"franchiseId": franchiseID,
		"name":        "Home Club " + label,
		"slug":        "home-club-" + slug,
	}, stdhttp.StatusCreated)
	homeClubID := cujID(t, homeClub, "id")

	awayClub := a.post(t, "/v1/clubs", map[string]any{
		"franchiseId": franchiseID,
		"name":        "Away Club " + label,
		"slug":        "away-club-" + slug,
	}, stdhttp.StatusCreated)
	awayClubID := cujID(t, awayClub, "id")

	homeTeam := a.post(t, "/v1/teams", map[string]any{
		"clubId": homeClubID,
		"name":   "Home Team " + label,
		"slug":   "home-team-" + slug,
	}, stdhttp.StatusCreated)
	homeTeamID := cujID(t, homeTeam, "id")

	awayTeam := a.post(t, "/v1/teams", map[string]any{
		"clubId": awayClubID,
		"name":   "Away Team " + label,
		"slug":   "away-team-" + slug,
	}, stdhttp.StatusCreated)
	awayTeamID := cujID(t, awayTeam, "id")

	homePlayer := a.post(t, "/v1/players", map[string]any{
		"displayName": "Home Player " + label,
		"slug":        "home-player-" + slug,
	}, stdhttp.StatusCreated)
	homePlayerID := cujID(t, homePlayer, "id")

	awayPlayer := a.post(t, "/v1/players", map[string]any{
		"displayName": "Away Player " + label,
		"slug":        "away-player-" + slug,
	}, stdhttp.StatusCreated)
	awayPlayerID := cujID(t, awayPlayer, "id")

	a.post(t, "/v1/roster-memberships", map[string]any{
		"playerId": homePlayerID,
		"teamId":   homeTeamID,
	}, stdhttp.StatusCreated)
	a.post(t, "/v1/roster-memberships", map[string]any{
		"playerId": awayPlayerID,
		"teamId":   awayTeamID,
	}, stdhttp.StatusCreated)

	queue := a.post(t, "/v1/queues", map[string]any{
		"name": "CUJ Queue " + label,
		"slug": "scrim-3v3-" + slug,
	}, stdhttp.StatusCreated)
	queueID := cujID(t, queue, "id")

	a.post(t, "/v1/platform-accounts/link", map[string]any{
		"subject":             subject,
		"provider":            "steam",
		"providerAccountId":   "steam-" + slug,
		"providerAccountName": "Steam " + label,
	}, stdhttp.StatusCreated)

	return cujLeagueFixture{
		LeagueID:    leagueID,
		FranchiseID: franchiseID,
		HomeClubID:  homeClubID,
		AwayClubID:  awayClubID,
		HomeTeamID:  homeTeamID,
		AwayTeamID:  awayTeamID,
		HomePlayer:  homePlayerID,
		AwayPlayer:  awayPlayerID,
		QueueID:     queueID,
		Subject:     subject,
	}
}

func (a *cujApp) post(t *testing.T, path string, payload map[string]any, expectedStatus int) map[string]any {
	t.Helper()
	return a.postWithRole(t, path, payload, "", expectedStatus)
}

func (a *cujApp) postAsAdmin(t *testing.T, path string, payload map[string]any, expectedStatus int) map[string]any {
	t.Helper()
	return a.postWithRole(t, path, payload, "admin", expectedStatus)
}

func (a *cujApp) postWithRole(t *testing.T, path string, payload map[string]any, role string, expectedStatus int) map[string]any {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal %s payload: %v", path, err)
	}
	req := httptest.NewRequest(stdhttp.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if role != "" {
		req.Header.Set("Authorization", "Bearer local:cuj-"+role+":"+role)
	}

	return a.serveJSON(t, req, path, expectedStatus)
}

func (a *cujApp) patch(t *testing.T, path string, payload map[string]any, expectedStatus int) map[string]any {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal %s payload: %v", path, err)
	}
	req := httptest.NewRequest(stdhttp.MethodPatch, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return a.serveJSON(t, req, path, expectedStatus)
}

func (a *cujApp) delete(t *testing.T, path string, payload map[string]any, expectedStatus int) map[string]any {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal %s payload: %v", path, err)
	}
	req := httptest.NewRequest(stdhttp.MethodDelete, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return a.serveJSON(t, req, path, expectedStatus)
}

func (a *cujApp) getObject(t *testing.T, path string, expectedStatus int) map[string]any {
	t.Helper()

	req := httptest.NewRequest(stdhttp.MethodGet, path, nil)
	return a.serveJSON(t, req, path, expectedStatus)
}

func (a *cujApp) getList(t *testing.T, path string, expectedStatus int) []map[string]any {
	t.Helper()

	req := httptest.NewRequest(stdhttp.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	a.server.Handler().ServeHTTP(rr, req)
	if rr.Code != expectedStatus {
		t.Fatalf("expected status %d for %s, got %d body=%s", expectedStatus, path, rr.Code, rr.Body.String())
	}
	if expectedStatus >= 400 {
		return nil
	}
	var data []map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode %s response: %v", path, err)
	}
	return data
}

func (a *cujApp) serveJSON(t *testing.T, req *stdhttp.Request, path string, expectedStatus int) map[string]any {
	t.Helper()

	rr := httptest.NewRecorder()
	a.server.Handler().ServeHTTP(rr, req)
	if rr.Code != expectedStatus {
		t.Fatalf("expected status %d for %s, got %d body=%s", expectedStatus, path, rr.Code, rr.Body.String())
	}
	if expectedStatus >= 400 {
		return nil
	}
	var data map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode %s response: %v", path, err)
	}
	return data
}

func cujID(t *testing.T, data map[string]any, key string) int64 {
	t.Helper()
	value, ok := data[key]
	if !ok {
		t.Fatalf("expected response key %q in %#v", key, data)
	}
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int64:
		return typed
	case int:
		return int64(typed)
	default:
		t.Fatalf("expected numeric response key %q, got %T", key, value)
		return 0
	}
}

func cujString(t *testing.T, data map[string]any, key string) string {
	t.Helper()
	value, ok := data[key]
	if !ok {
		t.Fatalf("expected response key %q in %#v", key, data)
	}
	typed, ok := value.(string)
	if !ok {
		t.Fatalf("expected string response key %q, got %T", key, value)
	}
	return typed
}
