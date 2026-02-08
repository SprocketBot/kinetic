package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
	"github.com/sprocketbot/sprocket-v3/internal/platform/config"
)

type probeResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type adminPingResponse struct {
	Status  string `json:"status"`
	Subject string `json:"subject"`
}

type fakeHierarchyStore struct {
	leagueToReturn     hierarchy.League
	franchiseToReturn  hierarchy.Franchise
	clubToReturn       hierarchy.Club
	leaguesToList      []hierarchy.League
	franchisesToList   []hierarchy.Franchise
	clubsToList        []hierarchy.Club
	createLeagueErr    error
	createFranchiseErr error
	createClubErr      error
}

func (f *fakeHierarchyStore) CreateLeague(_ context.Context, _ hierarchy.CreateLeagueInput) (hierarchy.League, error) {
	return f.leagueToReturn, f.createLeagueErr
}
func (f *fakeHierarchyStore) ListLeagues(_ context.Context) ([]hierarchy.League, error) {
	return f.leaguesToList, nil
}
func (f *fakeHierarchyStore) CreateFranchise(_ context.Context, _ hierarchy.CreateFranchiseInput) (hierarchy.Franchise, error) {
	return f.franchiseToReturn, f.createFranchiseErr
}
func (f *fakeHierarchyStore) ListFranchises(_ context.Context) ([]hierarchy.Franchise, error) {
	return f.franchisesToList, nil
}
func (f *fakeHierarchyStore) CreateClub(_ context.Context, _ hierarchy.CreateClubInput) (hierarchy.Club, error) {
	return f.clubToReturn, f.createClubErr
}
func (f *fakeHierarchyStore) ListClubs(_ context.Context) ([]hierarchy.Club, error) {
	return f.clubsToList, nil
}

func TestHealthEndpoint(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	assertProbePayload(t, rr.Body, "ok")
}

func TestReadyEndpoint(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	assertProbePayload(t, rr.Body, "ready")
}

func TestAdminPingRequiresAuthentication(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/ping", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestAdminPingRejectsInsufficientRole(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/ping", nil)
	req.Header.Set("Authorization", "Bearer local:bob:observer")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestAdminPingAllowsAdmin(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/ping", nil)
	req.Header.Set("Authorization", "Bearer local:alice:admin")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var payload adminPingResponse
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode admin ping payload: %v", err)
	}

	if payload.Subject != "alice" {
		t.Fatalf("expected subject alice, got %s", payload.Subject)
	}
}

func TestLeaguesEndpointUnavailableWithoutStore(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/leagues", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rr.Code)
	}
}

func TestCreateLeagueSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		leagueToReturn: hierarchy.League{
			ID:        1,
			Name:      "MLE",
			Slug:      "mle",
			IsActive:  true,
			CreatedAt: now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/leagues", strings.NewReader(`{"name":"MLE","slug":"mle"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestCreateLeagueValidationError(t *testing.T) {
	store := &fakeHierarchyStore{
		createLeagueErr: hierarchy.ErrInvalidInput,
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/leagues", strings.NewReader(`{"name":"MLE","slug":"BAD-SLUG"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func assertProbePayload(t *testing.T, body io.Reader, expectedStatus string) {
	t.Helper()

	var payload probeResponse
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode JSON payload: %v", err)
	}

	if payload.Status != expectedStatus {
		t.Fatalf("expected status %q, got %q", expectedStatus, payload.Status)
	}

	if payload.Service != "sprocket-v3-api" {
		t.Fatalf("unexpected service value %q", payload.Service)
	}
}
