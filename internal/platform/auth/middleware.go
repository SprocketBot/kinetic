package auth

import (
	"net/http"

	"github.com/kineticbot/kinetic-v3/internal/domain/authz"
)

func Authentication(validator TokenValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := validator.Validate(r.Header.Get("Authorization"))
		if err != nil {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), principal)))
	})
}

func RequirePermission(evaluator authz.Evaluator, resource, action string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal := PrincipalFromContext(r.Context())
		if principal.Anonymous {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		if !evaluator.Allowed(principal.Roles, resource, action) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
