package http

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
	"github.com/kineticbot/kinetic-v3/internal/domain/skillgroup"
)

func TestJSONHelpers(t *testing.T) {
	response := httptest.NewRecorder()
	writeJSON(response, http.StatusCreated, map[string]string{"ok": "yes"})
	if response.Code != http.StatusCreated || response.Header().Get("Content-Type") != "application/json" || response.Body.String() != "{\"ok\":\"yes\"}\n" {
		t.Fatalf("response = %#v body=%q", response.Result(), response.Body.String())
	}

	var payload struct {
		Name string `json:"name"`
	}
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"test"}`))
	if err := decodeJSON(request, &payload); err != nil || payload.Name != "test" {
		t.Fatalf("payload=%#v err=%v", payload, err)
	}
	request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"unknown":true}`))
	if err := decodeJSON(request, &payload); err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestRequestParsingHelpers(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/things/42?leagueId=7", nil)
	request.SetPathValue("id", "42")
	if got, err := parsePathID(request); err != nil || got != 42 {
		t.Fatalf("parsePathID=%d, %v", got, err)
	}
	if got, err := parseIntQueryParam(request, "leagueId"); err != nil || got != 7 {
		t.Fatalf("parseIntQueryParam=%d, %v", got, err)
	}
	if _, err := parseIntQueryParam(request, "missing"); err == nil {
		t.Fatal("expected missing query error")
	}
	if got := firstNonEmpty(" ", " first ", "second"); got != "first" {
		t.Fatalf("firstNonEmpty=%q", got)
	}
	if got := firstNonEmpty(" ", ""); got != "" {
		t.Fatalf("firstNonEmpty=%q", got)
	}
}

func TestErrorMappings(t *testing.T) {
	tests := []struct {
		name   string
		handle func(http.ResponseWriter, error)
		err    error
		want   int
	}{
		{"hierarchy invalid", handleHierarchyError, hierarchy.ErrInvalidInput, http.StatusBadRequest},
		{"hierarchy conflict", handleHierarchyError, hierarchy.ErrConflict, http.StatusConflict},
		{"hierarchy dependency", handleHierarchyError, hierarchy.ErrDependency, http.StatusConflict},
		{"hierarchy internal", handleHierarchyError, errors.New("boom"), http.StatusInternalServerError},
		{"skill not found", handleSkillGroupError, skillgroup.ErrNotFound, http.StatusNotFound},
		{"skill invalid", handleSkillGroupError, skillgroup.ErrInvalidInput, http.StatusBadRequest},
		{"skill conflict", handleSkillGroupError, skillgroup.ErrConflict, http.StatusConflict},
		{"skill dependency", handleSkillGroupError, skillgroup.ErrDependency, http.StatusConflict},
		{"skill internal", handleSkillGroupError, errors.New("boom"), http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			tt.handle(response, tt.err)
			if response.Code != tt.want || !strings.HasSuffix(response.Body.String(), "\n") {
				t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
			}
		})
	}
}
