package http

import (
	"github.com/sprocketbot/sprocket-v3/internal/platform/auth"
	"net/http"
	"strings"
	"time"
)

func (r routeRegistrar) registerAuthRoutes(mux *http.ServeMux) {
	logger := r.logger
	tokenValidator := r.tokenValidator
	evaluator := r.evaluator
	sessionTTL := r.sessionTTL
	sessionCookieName := r.sessionCookieName
	sessionSecret := r.sessionSecret
	authLocalLogin := r.authLocalLogin
	webBaseURL := r.webBaseURL
	discordClientID := r.discordClientID
	discordClientSecret := r.discordClientSecret
	discordRedirectURL := r.discordRedirectURL

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{
			Status:    "ok",
			Service:   "sprocket-v3-api",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{
			Status:    "ready",
			Service:   "sprocket-v3-api",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	})

	mux.HandleFunc("/v1/session", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sessionPrincipal, ok := readSessionPrincipal(r, sessionCookieName, sessionSecret)
			if !ok {
				http.Error(w, "not authenticated", http.StatusUnauthorized)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"subject":     sessionPrincipal.Subject,
				"displayName": sessionPrincipal.DisplayName,
				"roles":       sessionPrincipal.Roles,
			})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/auth/providers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		discordEnabled := discordClientID != "" && discordClientSecret != ""
		type providerInfo struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Enabled bool   `json:"enabled"`
		}
		providers := []providerInfo{
			{ID: "discord", Name: "Discord", Enabled: discordEnabled},
		}
		writeJSON(w, http.StatusOK, providers)
	})

	mux.HandleFunc("/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		provider := strings.TrimSpace(r.URL.Query().Get("provider"))
		switch provider {
		case "discord":
			if discordClientID == "" || discordClientSecret == "" {
				http.Error(w, "discord oauth not configured", http.StatusNotImplemented)
				return
			}
			state, err := generateOAuthState()
			if err != nil {
				http.Error(w, "failed to generate state", http.StatusInternalServerError)
				return
			}
			setOAuthStateCookie(w, state)
			redirectURL := discordRedirectURL
			if redirectURL == "" {
				redirectURL = strings.TrimRight(webBaseURL, "/") + "/v1/auth/callback?provider=discord"
			}
			http.Redirect(w, r, discordBuildAuthorizeURL(discordClientID, redirectURL, state), http.StatusFound)
		case "", "local":
			if !authLocalLogin {
				http.Error(w, "local authentication disabled", http.StatusNotFound)
				return
			}
			// Local/dev flow: redirect straight to callback with query params forwarded.
			callbackURL := "/v1/auth/callback?" + buildAuthCallbackQuery(r.URL.Query(), webBaseURL)
			http.Redirect(w, r, callbackURL, http.StatusFound)
		default:
			http.Error(w, "unknown provider", http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/v1/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		provider := strings.TrimSpace(r.URL.Query().Get("provider"))
		query := r.URL.Query()

		switch provider {
		case "discord":
			// Validate state parameter against the state cookie.
			providedState := query.Get("state")
			if !validateOAuthStateCookie(r, providedState) {
				http.Error(w, "invalid oauth state", http.StatusBadRequest)
				return
			}
			// Clear the state cookie.
			http.SetCookie(w, &http.Cookie{
				Name:     oauthStateCookieName,
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})

			code := query.Get("code")
			if code == "" {
				http.Error(w, "missing code parameter", http.StatusBadRequest)
				return
			}

			redirectURL := discordRedirectURL
			if redirectURL == "" {
				redirectURL = strings.TrimRight(webBaseURL, "/") + "/v1/auth/callback?provider=discord"
			}

			accessToken, err := discordExchangeCode(r.Context(), discordClientID, discordClientSecret, redirectURL, code)
			if err != nil {
				logger.Error("discord code exchange failed", "error", err)
				http.Error(w, "discord authentication failed", http.StatusBadGateway)
				return
			}

			discordID, discordUsername, err := discordGetCurrentUser(r.Context(), accessToken)
			if err != nil {
				logger.Error("discord user fetch failed", "error", err)
				http.Error(w, "discord authentication failed", http.StatusBadGateway)
				return
			}

			principal := auth.SessionPrincipal{
				Subject:     "discord:" + discordID,
				DisplayName: discordUsername,
				Roles:       []string{"player"},
				ExpiresAt:   time.Now().UTC().Add(sessionTTL),
			}

			token, err := auth.SignSessionToken(principal, sessionSecret)
			if err != nil {
				http.Error(w, "failed to issue session", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    token,
				Path:     "/",
				Expires:  principal.ExpiresAt,
				MaxAge:   int(sessionTTL.Seconds()),
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			http.Redirect(w, r, normalizeRedirectURL(query.Get("redirect"), webBaseURL), http.StatusFound)

		case "", "local":
			if !authLocalLogin {
				http.Error(w, "local authentication disabled", http.StatusNotFound)
				return
			}
			// Local/dev flow: build session directly from query parameters.
			principal := auth.SessionPrincipal{
				Subject:     firstNonEmpty(strings.TrimSpace(query.Get("subject")), "local-player"),
				DisplayName: firstNonEmpty(strings.TrimSpace(query.Get("displayName")), strings.TrimSpace(query.Get("subject"))),
				Roles:       parseRoleList(query.Get("roles")),
				ExpiresAt:   time.Now().UTC().Add(sessionTTL),
			}
			if principal.DisplayName == "" {
				principal.DisplayName = principal.Subject
			}

			token, err := auth.SignSessionToken(principal, sessionSecret)
			if err != nil {
				http.Error(w, "failed to issue session", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    token,
				Path:     "/",
				Expires:  principal.ExpiresAt,
				MaxAge:   int(sessionTTL.Seconds()),
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			http.Redirect(w, r, normalizeRedirectURL(query.Get("redirect"), webBaseURL), http.StatusFound)
		default:
			http.Error(w, "unknown provider", http.StatusBadRequest)
		}
	})

	mux.HandleFunc("/v1/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookieName,
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	adminPingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal := auth.PrincipalFromContext(r.Context())
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"subject": principal.Subject,
		})
	})
	mux.Handle("/v1/admin/ping", auth.Authentication(
		tokenValidator,
		auth.RequirePermission(evaluator, "admin.ping", "read", adminPingHandler),
	))
}
