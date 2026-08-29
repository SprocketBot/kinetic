package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kineticbot/kinetic-v3/internal/domain/authz"
)

type tokenValidatorFunc func(string) (Principal, error)

func (f tokenValidatorFunc) Validate(header string) (Principal, error) { return f(header) }

func TestAuthentication(t *testing.T) {
	t.Run("rejects invalid token", func(t *testing.T) {
		nextCalled := false
		handler := Authentication(tokenValidatorFunc(func(string) (Principal, error) {
			return Principal{}, errors.New("bad token")
		}), http.HandlerFunc(func(http.ResponseWriter, *http.Request) { nextCalled = true }))

		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

		if response.Code != http.StatusUnauthorized || nextCalled {
			t.Fatalf("status=%d nextCalled=%v", response.Code, nextCalled)
		}
	})

	t.Run("adds principal to context", func(t *testing.T) {
		want := NewPrincipal("player-1", []string{"player"})
		handler := Authentication(tokenValidatorFunc(func(header string) (Principal, error) {
			if header != "Bearer token" {
				t.Fatalf("header = %q", header)
			}
			return want, nil
		}), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if got := PrincipalFromContext(r.Context()); got.Subject != want.Subject {
				t.Fatalf("principal = %#v", got)
			}
			w.WriteHeader(http.StatusNoContent)
		}))

		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Authorization", "Bearer token")
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d", response.Code)
		}
	})
}

func TestRequirePermission(t *testing.T) {
	evaluator := authz.NewStaticEvaluator([]authz.Permission{{Role: "reader", Resource: "thing", Action: "read"}})
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	handler := RequirePermission(evaluator, "thing", "read", next)

	tests := []struct {
		name      string
		principal Principal
		want      int
	}{
		{name: "anonymous", principal: AnonymousPrincipal(), want: http.StatusUnauthorized},
		{name: "forbidden", principal: NewPrincipal("p", []string{"writer"}), want: http.StatusForbidden},
		{name: "allowed", principal: NewPrincipal("p", []string{"reader"}), want: http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request = request.WithContext(WithPrincipal(request.Context(), tt.principal))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tt.want {
				t.Fatalf("status=%d, want %d", response.Code, tt.want)
			}
		})
	}
}
