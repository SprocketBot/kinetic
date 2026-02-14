package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
	"github.com/sprocketbot/sprocket-v3/internal/platform/config"
	"github.com/sprocketbot/sprocket-v3/internal/platform/db"
)

func TestHierarchyAPICreateAndConstraints(t *testing.T) {
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	conn, err := db.Open(testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := db.Ping(ctx, conn); err != nil {
		t.Fatalf("failed to ping test DB: %v", err)
	}

	migrator := db.NewMigrator(conn, "../../../migrations")
	if _, err := migrator.Up(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	store := db.NewHierarchyStore(conn)
	server := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	suffix := time.Now().UnixNano()
	leagueName := fmt.Sprintf("League %d", suffix)
	leagueSlug := fmt.Sprintf("league-%d", suffix)

	leagueResp := createEntity(t, server, "/v1/leagues", map[string]any{
		"name": leagueName,
		"slug": leagueSlug,
	}, http.StatusCreated)

	// Duplicate slug should produce conflict.
	createEntity(t, server, "/v1/leagues", map[string]any{
		"name": fmt.Sprintf("Other %d", suffix),
		"slug": leagueSlug,
	}, http.StatusConflict)

	leagueID := int64(leagueResp["id"].(float64))

	// Missing parent league should produce dependency conflict.
	createEntity(t, server, "/v1/franchises", map[string]any{
		"leagueId": int64(99999999),
		"name":     "Invalid Franchise",
		"slug":     fmt.Sprintf("invalid-franchise-%d", suffix),
	}, http.StatusConflict)

	franchiseResp := createEntity(t, server, "/v1/franchises", map[string]any{
		"leagueId": leagueID,
		"name":     fmt.Sprintf("Franchise %d", suffix),
		"slug":     fmt.Sprintf("franchise-%d", suffix),
	}, http.StatusCreated)
	franchiseID := int64(franchiseResp["id"].(float64))

	createEntity(t, server, "/v1/clubs", map[string]any{
		"franchiseId": int64(99999999),
		"name":        "Invalid Club",
		"slug":        fmt.Sprintf("invalid-club-%d", suffix),
	}, http.StatusConflict)

	clubResp := createEntity(t, server, "/v1/clubs", map[string]any{
		"franchiseId": franchiseID,
		"name":        fmt.Sprintf("Club %d", suffix),
		"slug":        fmt.Sprintf("club-%d", suffix),
	}, http.StatusCreated)
	clubID := int64(clubResp["id"].(float64))

	clubTwoResp := createEntity(t, server, "/v1/clubs", map[string]any{
		"franchiseId": franchiseID,
		"name":        fmt.Sprintf("Club Two %d", suffix),
		"slug":        fmt.Sprintf("club-two-%d", suffix),
	}, http.StatusCreated)
	clubTwoID := int64(clubTwoResp["id"].(float64))

	createEntity(t, server, "/v1/teams", map[string]any{
		"clubId": int64(99999999),
		"name":   "Invalid Team",
		"slug":   fmt.Sprintf("invalid-team-%d", suffix),
	}, http.StatusConflict)

	teamResp := createEntity(t, server, "/v1/teams", map[string]any{
		"clubId": clubID,
		"name":   fmt.Sprintf("Team %d", suffix),
		"slug":   fmt.Sprintf("team-%d", suffix),
	}, http.StatusCreated)
	teamID := int64(teamResp["id"].(float64))

	teamTwoResp := createEntity(t, server, "/v1/teams", map[string]any{
		"clubId": clubID,
		"name":   fmt.Sprintf("Team Two %d", suffix),
		"slug":   fmt.Sprintf("team-two-%d", suffix),
	}, http.StatusCreated)
	teamTwoID := int64(teamTwoResp["id"].(float64))

	createEntity(t, server, "/v1/players", map[string]any{
		"displayName": fmt.Sprintf("Player %d", suffix),
		"slug":        fmt.Sprintf("player-%d", suffix),
	}, http.StatusCreated)

	playerTwoResp := createEntity(t, server, "/v1/players", map[string]any{
		"displayName": fmt.Sprintf("Player Two %d", suffix),
		"slug":        fmt.Sprintf("player-two-%d", suffix),
	}, http.StatusCreated)
	playerTwoID := int64(playerTwoResp["id"].(float64))

	createEntity(t, server, "/v1/roster-memberships", map[string]any{
		"playerId": playerTwoID,
		"teamId":   teamID,
	}, http.StatusCreated)

	createEntity(t, server, "/v1/roster-memberships", map[string]any{
		"playerId": playerTwoID,
		"teamId":   teamID,
	}, http.StatusConflict)

	createEntity(t, server, "/v1/roster-memberships", map[string]any{
		"playerId": playerTwoID,
		"teamId":   int64(99999999),
	}, http.StatusConflict)

	createEntity(t, server, "/v1/roster-memberships", map[string]any{
		"playerId": int64(99999999),
		"teamId":   teamID,
	}, http.StatusConflict)

	queueResp := createEntity(t, server, "/v1/queues", map[string]any{
		"name": fmt.Sprintf("Queue %d", suffix),
		"slug": fmt.Sprintf("queue-%d", suffix),
	}, http.StatusCreated)
	queueID := int64(queueResp["id"].(float64))

	createEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": queueID,
		"teamId":  teamID,
	}, http.StatusCreated)

	createEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": queueID,
		"teamId":  teamID,
	}, http.StatusConflict)

	createEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": queueID,
		"teamId":  int64(99999999),
	}, http.StatusConflict)

	createEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": int64(99999999),
		"teamId":  teamTwoID,
	}, http.StatusConflict)

	patchEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": queueID,
		"teamId":  teamID,
		"stage":   int32(2),
	}, http.StatusOK)

	patchEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": queueID,
		"teamId":  teamID,
		"stage":   int32(0),
	}, http.StatusBadRequest)

	deleteEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": queueID,
		"teamId":  teamID,
	}, http.StatusOK)

	deleteEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": queueID,
		"teamId":  teamID,
	}, http.StatusConflict)

	seasonResp := createEntity(t, server, "/v1/seasons", map[string]any{
		"name": fmt.Sprintf("Season %d", suffix),
		"slug": fmt.Sprintf("season-%d", suffix),
	}, http.StatusCreated)
	seasonID := int64(seasonResp["id"].(float64))

	groupResp := createEntity(t, server, "/v1/schedule-groups", map[string]any{
		"seasonId": seasonID,
		"name":     fmt.Sprintf("Week %d", suffix),
		"sequence": int32(1),
	}, http.StatusCreated)
	groupID := int64(groupResp["id"].(float64))

	createEntity(t, server, "/v1/fixtures", map[string]any{
		"scheduleGroupId": groupID,
		"homeClubId":      clubID,
		"awayClubId":      int64(99999999),
	}, http.StatusConflict)

	fixtureResp := createEntity(t, server, "/v1/fixtures", map[string]any{
		"scheduleGroupId": groupID,
		"homeClubId":      clubID,
		"awayClubId":      clubTwoID,
	}, http.StatusCreated)
	fixtureID := int64(fixtureResp["id"].(float64))

	createEntity(t, server, "/v1/matches", map[string]any{
		"fixtureId":    fixtureID,
		"homeTeamId":   teamID,
		"awayTeamId":   teamTwoID,
		"state":        "ready",
		"scheduledFor": nil,
	}, http.StatusBadRequest)

	createEntity(t, server, "/v1/matches", map[string]any{
		"fixtureId":  fixtureID,
		"homeTeamId": teamID,
		"awayTeamId": teamTwoID,
		"state":      "planned",
	}, http.StatusCreated)

	createEntity(t, server, "/v1/matches", map[string]any{
		"fixtureId":          fixtureID,
		"homeTeamId":         teamID,
		"awayTeamId":         teamTwoID,
		"state":              "ready",
		"scheduledFor":       "2030-01-01T10:00:00Z",
		"homeTimeRatifiedAt": "2030-01-01T08:00:00Z",
		"awayTimeRatifiedAt": "2030-01-01T09:00:00Z",
	}, http.StatusCreated)

	scrimResp := createEntity(t, server, "/v1/scrims", map[string]any{
		"queueId":    queueID,
		"homeTeamId": teamID,
		"awayTeamId": teamTwoID,
		"state":      "created",
	}, http.StatusCreated)
	scrimID := int64(scrimResp["id"].(float64))

	patchEntity(t, server, "/v1/scrims", map[string]any{
		"scrimId": scrimID,
		"state":   "in_progress",
	}, http.StatusOK)

	patchEntity(t, server, "/v1/scrims", map[string]any{
		"scrimId": scrimID,
		"state":   "closed",
	}, http.StatusOK)

	patchEntity(t, server, "/v1/scrims", map[string]any{
		"scrimId": scrimID,
		"state":   "in_progress",
	}, http.StatusConflict)

	if _, err := conn.ExecContext(
		ctx,
		`INSERT INTO player_ratings(player_id, context_key, rating, uncertainty, matches_played) VALUES ($1, $2, $3, $4, $5)`,
		playerTwoID,
		"scrim-3v3",
		1025,
		320,
		7,
	); err != nil {
		t.Fatalf("failed to insert player rating baseline row: %v", err)
	}

	createEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": queueID,
		"teamId":  teamID,
	}, http.StatusCreated)
	createEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": queueID,
		"teamId":  teamTwoID,
	}, http.StatusCreated)

	createEntity(t, server, "/v1/scrim-promotions", map[string]any{
		"queueId": queueID,
	}, http.StatusCreated)

	createEntity(t, server, "/v1/scrim-promotions", map[string]any{
		"queueId": queueID,
	}, http.StatusConflict)

	createEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": queueID,
		"teamId":  teamID,
	}, http.StatusCreated)

	assertListNotEmpty(t, server, "/v1/leagues")
	assertListNotEmpty(t, server, "/v1/franchises")
	assertListNotEmpty(t, server, "/v1/clubs")
	assertListNotEmpty(t, server, "/v1/teams")
	assertListNotEmpty(t, server, "/v1/players")
	assertListNotEmpty(t, server, "/v1/roster-memberships")
	assertListNotEmpty(t, server, "/v1/queues")
	assertListNotEmpty(t, server, "/v1/queue-entries")
	assertListNotEmpty(t, server, "/v1/scrims")
	assertListNotEmpty(t, server, "/v1/player-ratings")
	assertListNotEmpty(t, server, "/v1/matchmaking-decisions")
	assertListNotEmpty(t, server, "/v1/seasons")
	assertListNotEmpty(t, server, "/v1/schedule-groups")
	assertListNotEmpty(t, server, "/v1/fixtures")
	assertListNotEmpty(t, server, "/v1/matches")
}

func TestHierarchyAPIValidationFailure(t *testing.T) {
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	conn, err := db.Open(testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := db.Ping(ctx, conn); err != nil {
		t.Fatalf("failed to ping test DB: %v", err)
	}

	migrator := db.NewMigrator(conn, "../../../migrations")
	if _, err := migrator.Up(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	store := db.NewHierarchyStore(conn)
	server := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	createEntity(t, server, "/v1/leagues", map[string]any{
		"name": "Invalid League",
		"slug": "NOT-VALID",
	}, http.StatusBadRequest)

	createEntity(t, server, "/v1/roster-memberships", map[string]any{
		"playerId": int64(0),
		"teamId":   int64(1),
	}, http.StatusBadRequest)

	createEntity(t, server, "/v1/queues", map[string]any{
		"name": "Queue Validation",
		"slug": "NOT-VALID",
	}, http.StatusBadRequest)

	createEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": int64(0),
		"teamId":  int64(1),
	}, http.StatusBadRequest)

	patchEntity(t, server, "/v1/queue-entries", map[string]any{
		"queueId": int64(1),
		"teamId":  int64(1),
		"stage":   int32(0),
	}, http.StatusBadRequest)

	createEntity(t, server, "/v1/scrims", map[string]any{
		"queueId":    int64(1),
		"homeTeamId": int64(1),
		"awayTeamId": int64(1),
		"state":      "created",
	}, http.StatusBadRequest)

	patchEntity(t, server, "/v1/scrims", map[string]any{
		"scrimId": int64(1),
		"state":   "created",
	}, http.StatusBadRequest)

	createEntity(t, server, "/v1/scrim-promotions", map[string]any{
		"queueId": int64(0),
	}, http.StatusBadRequest)

	createEntity(t, server, "/v1/matches", map[string]any{
		"fixtureId":  int64(1),
		"homeTeamId": int64(1),
		"awayTeamId": int64(1),
		"state":      "planned",
	}, http.StatusBadRequest)
}

func createEntity(t *testing.T, server *Server, path string, payload map[string]any, expectedStatus int) map[string]any {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != expectedStatus {
		t.Fatalf("expected status %d for %s, got %d body=%s", expectedStatus, path, rr.Code, rr.Body.String())
	}

	if expectedStatus >= 400 {
		return nil
	}

	var data map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&data); err != nil {
		t.Fatalf("failed to decode response for %s: %v", path, err)
	}
	return data
}

func assertListNotEmpty(t *testing.T, server *Server, path string) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d for %s, got %d", http.StatusOK, path, rr.Code)
	}

	var items []map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&items); err != nil {
		t.Fatalf("failed to decode list response for %s: %v", path, err)
	}
	if len(items) == 0 {
		t.Fatalf("expected non-empty list for %s", path)
	}
}

func deleteEntity(t *testing.T, server *Server, path string, payload map[string]any, expectedStatus int) {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != expectedStatus {
		t.Fatalf("expected status %d for %s, got %d body=%s", expectedStatus, path, rr.Code, rr.Body.String())
	}
}

func patchEntity(t *testing.T, server *Server, path string, payload map[string]any, expectedStatus int) {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	server.Handler().ServeHTTP(rr, req)

	if rr.Code != expectedStatus {
		t.Fatalf("expected status %d for %s, got %d body=%s", expectedStatus, path, rr.Code, rr.Body.String())
	}
}

var _ hierarchy.Store = (*db.HierarchyStore)(nil)
