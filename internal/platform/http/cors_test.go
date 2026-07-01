package http

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sprocketbot/sprocket-v3/internal/platform/config"
)

func TestCORSPreflightAllowsConfiguredOrigin(t *testing.T) {
	srv := New(config.Config{
		Port:               "8080",
		LogLevel:           "info",
		WebBaseURL:         "https://sprocket.mlesports.gg",
		CORSAllowedOrigins: "https://sprocket.mlesports.gg",
		AuthLocalLogin:     true,
	}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodOptions, "/v1/session", nil)
	req.Header.Set("Origin", "https://sprocket.mlesports.gg")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rr := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected preflight status %d, got %d body=%s", http.StatusNoContent, rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://sprocket.mlesports.gg" {
		t.Fatalf("expected reflected CORS origin, got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentialed CORS, got %q", got)
	}
}

func TestCORSPreflightRejectsUnconfiguredOrigin(t *testing.T) {
	srv := New(config.Config{
		Port:               "8080",
		LogLevel:           "info",
		WebBaseURL:         "https://sprocket.mlesports.gg",
		CORSAllowedOrigins: "https://sprocket.mlesports.gg",
		AuthLocalLogin:     true,
	}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodOptions, "/v1/session", nil)
	req.Header.Set("Origin", "https://evil.example")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rr := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected preflight status %d, got %d body=%s", http.StatusForbidden, rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no CORS allow origin for rejected origin, got %q", got)
	}
}

func TestCORSActualResponseAllowsConfiguredOrigin(t *testing.T) {
	srv := New(config.Config{
		Port:               "8080",
		LogLevel:           "info",
		WebBaseURL:         "https://sprocket.mlesports.gg",
		CORSAllowedOrigins: "https://sprocket.mlesports.gg",
		AuthLocalLogin:     true,
	}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/session", nil)
	req.Header.Set("Origin", "https://sprocket.mlesports.gg")
	rr := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected session status %d, got %d body=%s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://sprocket.mlesports.gg" {
		t.Fatalf("expected reflected CORS origin, got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("expected credentialed CORS, got %q", got)
	}
}

func TestLocalAuthCanBeDisabled(t *testing.T) {
	srv := New(config.Config{
		Port:              "8080",
		LogLevel:          "info",
		WebBaseURL:        "https://sprocket.mlesports.gg",
		AuthLocalLogin:    false,
		AuthLocalLoginSet: true,
	}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/login?subject=alice", nil)
	rr := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected local auth disabled status %d, got %d body=%s", http.StatusNotFound, rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "local authentication disabled") {
		t.Fatalf("expected disabled local auth response, got %q", rr.Body.String())
	}
}

func TestUnknownAuthCallbackProviderIsRejected(t *testing.T) {
	srv := New(config.Config{
		Port:           "8080",
		LogLevel:       "info",
		WebBaseURL:     "http://localhost:5173",
		AuthLocalLogin: true,
	}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/callback?provider=github&subject=alice", nil)
	rr := httptest.NewRecorder()

	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected unknown provider status %d, got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}
