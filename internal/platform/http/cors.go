package http

import (
	"net/http"
	"strings"
)

func parseCORSAllowedOrigins(raw, fallback string) []string {
	source := strings.TrimSpace(raw)
	if source == "" {
		source = strings.TrimSpace(fallback)
	}

	seen := map[string]struct{}{}
	origins := make([]string, 0)
	for _, value := range strings.Split(source, ",") {
		origin := strings.TrimRight(strings.TrimSpace(value), "/")
		if origin == "" {
			continue
		}
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	return origins
}

func corsMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := map[string]struct{}{}
	allowAny := false
	for _, origin := range allowedOrigins {
		if origin == "*" {
			allowAny = true
			continue
		}
		allowed[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := strings.TrimRight(strings.TrimSpace(r.Header.Get("Origin")), "/")
		originAllowed := false
		if origin != "" {
			if allowAny {
				originAllowed = true
			} else if _, ok := allowed[origin]; ok {
				originAllowed = true
			}
		}

		if originAllowed {
			headers := w.Header()
			headers.Set("Access-Control-Allow-Origin", origin)
			headers.Set("Access-Control-Allow-Credentials", "true")
			headers.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			headers.Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
			headers.Add("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			if origin == "" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if !originAllowed {
				http.Error(w, "cors origin not allowed", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
