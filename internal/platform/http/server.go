package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sprocketbot/sprocket-v3/internal/domain/apitoken"
	"github.com/sprocketbot/sprocket-v3/internal/domain/authz"
	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
	"github.com/sprocketbot/sprocket-v3/internal/domain/notifications"
	"github.com/sprocketbot/sprocket-v3/internal/domain/orgconfig"
	"github.com/sprocketbot/sprocket-v3/internal/domain/replaystats"
	"github.com/sprocketbot/sprocket-v3/internal/domain/skillgroup"
	"github.com/sprocketbot/sprocket-v3/internal/platform/auth"
	"github.com/sprocketbot/sprocket-v3/internal/platform/config"
)

type Server struct {
	logger *slog.Logger
	http   *http.Server
}

type Dependencies struct {
	TokenValidator     auth.TokenValidator
	Evaluator          authz.Evaluator
	HierarchyStore     hierarchy.Store
	SkillGroupStore    skillgroup.Store
	OrgConfigStore     orgconfig.Store
	NotificationsStore notifications.Store
	APITokenStore      apitoken.Store
	ReplayStatsStore   replaystats.Store
}

type healthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

func New(cfg config.Config, logger *slog.Logger, deps Dependencies) *Server {
	mux := http.NewServeMux()
	tokenValidator := deps.TokenValidator
	if tokenValidator == nil {
		defaultValidator := auth.NewLocalTokenValidator("local")
		tokenValidator = defaultValidator
	}
	sessionTTL := parseSessionTTL(cfg.AuthSessionTTL)
	sessionCookieName := strings.TrimSpace(cfg.AuthSessionCookie)
	if sessionCookieName == "" {
		sessionCookieName = "sprocket_session"
	}
	sessionSecret := strings.TrimSpace(cfg.AuthSessionSecret)
	if sessionSecret == "" {
		sessionSecret = "dev-insecure-session-secret"
	}
	webBaseURL := strings.TrimRight(strings.TrimSpace(cfg.WebBaseURL), "/")
	if webBaseURL == "" {
		webBaseURL = "http://localhost:5173"
	}

	evaluator := deps.Evaluator
	if evaluator == nil {
		defaultEvaluator := authz.NewStaticEvaluator(authz.DefaultPermissions())
		evaluator = defaultEvaluator
	}

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

	discordClientID := strings.TrimSpace(cfg.DiscordClientID)
	discordClientSecret := strings.TrimSpace(cfg.DiscordClientSecret)
	discordRedirectURL := strings.TrimSpace(cfg.DiscordRedirectURL)

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

		default:
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

	mux.HandleFunc("/v1/leagues", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			leagues, err := deps.HierarchyStore.ListLeagues(r.Context())
			if err != nil {
				http.Error(w, "failed to list leagues", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, leagues)
		case http.MethodPost:
			var input hierarchy.CreateLeagueInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			league, err := deps.HierarchyStore.CreateLeague(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, league)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/franchises", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			franchises, err := deps.HierarchyStore.ListFranchises(r.Context())
			if err != nil {
				http.Error(w, "failed to list franchises", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, franchises)
		case http.MethodPost:
			var input hierarchy.CreateFranchiseInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			franchise, err := deps.HierarchyStore.CreateFranchise(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, franchise)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/clubs", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			clubs, err := deps.HierarchyStore.ListClubs(r.Context())
			if err != nil {
				http.Error(w, "failed to list clubs", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, clubs)
		case http.MethodPost:
			var input hierarchy.CreateClubInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			club, err := deps.HierarchyStore.CreateClub(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, club)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/teams", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			teams, err := deps.HierarchyStore.ListTeams(r.Context())
			if err != nil {
				http.Error(w, "failed to list teams", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, teams)
		case http.MethodPost:
			var input hierarchy.CreateTeamInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			team, err := deps.HierarchyStore.CreateTeam(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, team)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/players", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			players, err := deps.HierarchyStore.ListPlayers(r.Context())
			if err != nil {
				http.Error(w, "failed to list players", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, players)
		case http.MethodPost:
			var input hierarchy.CreatePlayerInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			player, err := deps.HierarchyStore.CreatePlayer(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, player)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/roster-memberships", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			memberships, err := deps.HierarchyStore.ListRosterMemberships(r.Context())
			if err != nil {
				http.Error(w, "failed to list roster memberships", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, memberships)
		case http.MethodPost:
			var input hierarchy.CreateRosterMembershipInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			membership, err := deps.HierarchyStore.CreateRosterMembership(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, membership)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/role-assignments", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			assignments, err := deps.HierarchyStore.ListRoleAssignments(r.Context())
			if err != nil {
				http.Error(w, "failed to list role assignments", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, assignments)
		case http.MethodPost:
			if !checkPermission(w, r, sessionCookieName, sessionSecret, tokenValidator, deps.APITokenStore, evaluator, authz.ResourceRoleAssignment, authz.ActionCreate) {
				return
			}
			var input hierarchy.AssignRoleInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			assignment, err := deps.HierarchyStore.AssignRole(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, assignment)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/role-assignments/revoke", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodPost:
			if !checkPermission(w, r, sessionCookieName, sessionSecret, tokenValidator, deps.APITokenStore, evaluator, authz.ResourceRoleAssignment, authz.ActionRevoke) {
				return
			}
			var input hierarchy.RevokeRoleInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			assignment, err := deps.HierarchyStore.RevokeRole(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, assignment)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/queues", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			queues, err := deps.HierarchyStore.ListQueues(r.Context())
			if err != nil {
				http.Error(w, "failed to list queues", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, queues)
		case http.MethodPost:
			var input hierarchy.CreateQueueInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			queue, err := deps.HierarchyStore.CreateQueue(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, queue)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/platform-accounts", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			subject := strings.TrimSpace(r.URL.Query().Get("subject"))
			if subject == "" {
				if sessionPrincipal, ok := readSessionPrincipal(r, sessionCookieName, sessionSecret); ok {
					subject = sessionPrincipal.Subject
				}
			}
			if subject == "" {
				http.Error(w, "subject is required", http.StatusBadRequest)
				return
			}
			links, err := deps.HierarchyStore.ListPlatformAccountLinks(r.Context(), subject)
			if err != nil {
				http.Error(w, "failed to list platform account links", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, links)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/platform-accounts/link", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodPost:
			var input hierarchy.LinkPlatformAccountInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(input.Subject) == "" {
				if sessionPrincipal, ok := readSessionPrincipal(r, sessionCookieName, sessionSecret); ok {
					input.Subject = sessionPrincipal.Subject
				}
			}
			link, err := deps.HierarchyStore.LinkPlatformAccount(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, link)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/platform-accounts/unlink", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodPost:
			var input hierarchy.UnlinkPlatformAccountInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			if strings.TrimSpace(input.Subject) == "" {
				if sessionPrincipal, ok := readSessionPrincipal(r, sessionCookieName, sessionSecret); ok {
					input.Subject = sessionPrincipal.Subject
				}
			}
			link, err := deps.HierarchyStore.UnlinkPlatformAccount(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, link)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/eligibility", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			subject := strings.TrimSpace(r.URL.Query().Get("subject"))
			if subject == "" {
				if sessionPrincipal, ok := readSessionPrincipal(r, sessionCookieName, sessionSecret); ok {
					subject = sessionPrincipal.Subject
				}
			}
			if subject == "" {
				http.Error(w, "subject is required", http.StatusBadRequest)
				return
			}

			status, err := deps.HierarchyStore.GetEligibilityStatus(r.Context(), subject)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, status)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/queue-entries", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			entries, err := deps.HierarchyStore.ListActiveQueueEntries(r.Context())
			if err != nil {
				http.Error(w, "failed to list queue entries", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, entries)
		case http.MethodPost:
			var input hierarchy.EnqueueTeamInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			entry, err := deps.HierarchyStore.EnqueueTeam(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, entry)
		case http.MethodDelete:
			var input hierarchy.LeaveQueueInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			entry, err := deps.HierarchyStore.LeaveQueue(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, entry)
		case http.MethodPatch:
			var input hierarchy.AdvanceQueueEntryStageInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			entry, err := deps.HierarchyStore.AdvanceQueueEntryStage(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, entry)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/queue-bans", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			bans, err := deps.HierarchyStore.ListQueueBans(r.Context())
			if err != nil {
				http.Error(w, "failed to list queue bans", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, bans)
		case http.MethodPost:
			var input hierarchy.BanPlayerFromQueueInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			ban, err := deps.HierarchyStore.BanPlayerFromQueue(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, ban)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/queue-bans/lift", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodPost:
			var input hierarchy.UnbanPlayerFromQueueInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			ban, err := deps.HierarchyStore.UnbanPlayerFromQueue(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, ban)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/scrims", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			scrims, err := deps.HierarchyStore.ListScrims(r.Context())
			if err != nil {
				http.Error(w, "failed to list scrims", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, scrims)
		case http.MethodPost:
			var input hierarchy.CreateScrimInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			scrim, err := deps.HierarchyStore.CreateScrim(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, scrim)
		case http.MethodPatch:
			var input hierarchy.UpdateScrimStateInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			scrim, err := deps.HierarchyStore.UpdateScrimState(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, scrim)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/scrim-promotions", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodPost:
			if !checkPermission(w, r, sessionCookieName, sessionSecret, tokenValidator, deps.APITokenStore, evaluator, authz.ResourceScrimPromotion, authz.ActionCreate) {
				return
			}
			var input hierarchy.PromoteQueueToScrimInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			scrim, err := deps.HierarchyStore.PromoteQueueToScrim(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			startPopTimeoutWatcher(scrim.ID, 5*time.Minute, deps.HierarchyStore, logger)
			writeJSON(w, http.StatusCreated, scrim)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/scrim-promotions/process", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodPost:
			if !checkPermission(w, r, sessionCookieName, sessionSecret, tokenValidator, deps.APITokenStore, evaluator, authz.ResourceScrimPromotion, authz.ActionCreate) {
				return
			}
			var input hierarchy.ProcessQueuePromotionsInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			result, err := deps.HierarchyStore.ProcessQueuePromotions(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			logger.Info(
				"scrim promotion processing completed",
				"queueId", input.QueueID,
				"processedQueues", result.ProcessedQueues,
				"promotionsCreated", result.PromotionsCreated,
				"conflicts", result.Conflicts,
				"poppedScrims", len(result.PoppedScrimIDs),
			)
			for _, scrimID := range result.PoppedScrimIDs {
				startPopTimeoutWatcher(scrimID, 5*time.Minute, deps.HierarchyStore, logger)
			}
			writeJSON(w, http.StatusOK, result)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/promotion-processing-runs", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			runs, err := deps.HierarchyStore.ListPromotionProcessingRuns(r.Context())
			if err != nil {
				http.Error(w, "failed to list promotion processing runs", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, runs)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/player-ratings", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			ratings, err := deps.HierarchyStore.ListPlayerRatings(r.Context())
			if err != nil {
				http.Error(w, "failed to list player ratings", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, ratings)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/player-ratings/adjust", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodPost:
			var input hierarchy.AdjustPlayerRatingInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			rating, err := deps.HierarchyStore.AdjustPlayerRating(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, rating)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/rating-adjustments", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			adjustments, err := deps.HierarchyStore.ListRatingAdjustments(r.Context())
			if err != nil {
				http.Error(w, "failed to list rating adjustments", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, adjustments)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/matchmaking-decisions", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			decisions, err := deps.HierarchyStore.ListMatchmakingDecisions(r.Context())
			if err != nil {
				http.Error(w, "failed to list matchmaking decisions", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, decisions)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/result-submissions", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			submissions, err := deps.HierarchyStore.ListResultSubmissions(r.Context())
			if err != nil {
				http.Error(w, "failed to list result submissions", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, submissions)
		case http.MethodPost:
			var input hierarchy.CreateResultSubmissionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			submission, err := deps.HierarchyStore.CreateResultSubmission(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, submission)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/result-overrides", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			overrides, err := deps.HierarchyStore.ListResultOverrides(r.Context())
			if err != nil {
				http.Error(w, "failed to list result overrides", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, overrides)
		case http.MethodPost:
			if !checkPermission(w, r, sessionCookieName, sessionSecret, tokenValidator, deps.APITokenStore, evaluator, authz.ResourceResultOverride, authz.ActionCreate) {
				return
			}
			var input hierarchy.OverrideResultSubmissionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			submission, err := deps.HierarchyStore.OverrideResultSubmission(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, submission)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/result-submission-ratifications", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodPost:
			var input hierarchy.RatifyResultSubmissionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			submission, err := deps.HierarchyStore.RatifyResultSubmission(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, submission)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/result-submission-rejections", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodPost:
			var input hierarchy.RejectResultSubmissionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			submission, err := deps.HierarchyStore.RejectResultSubmission(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, submission)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/replay-evidence", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			evidence, err := deps.HierarchyStore.ListReplayEvidence(r.Context())
			if err != nil {
				http.Error(w, "failed to list replay evidence", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, evidence)
		case http.MethodPost:
			var input hierarchy.IngestReplayEvidenceInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			result, err := deps.HierarchyStore.IngestReplayEvidence(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			// Trigger async stub replay parse in the background.
			go func() {
				bgCtx := context.Background()
				if triggerErr := deps.HierarchyStore.TriggerReplayParse(bgCtx, result.Evidence.ID, input.ContextID, input.ContextType); triggerErr != nil {
					// Log only — the HTTP response has already been sent.
					_ = triggerErr
				}
			}()
			status := http.StatusCreated
			if result.Duplicate {
				status = http.StatusOK
			}
			writeJSON(w, status, result)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/replay-parse-runs", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			runs, err := deps.HierarchyStore.ListReplayParseRuns(r.Context())
			if err != nil {
				http.Error(w, "failed to list replay parse runs", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, runs)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/result-submission-replay-links", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			links, err := deps.HierarchyStore.ListResultSubmissionReplayLinks(r.Context())
			if err != nil {
				http.Error(w, "failed to list result submission replay links", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, links)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/result-submissions/{id}/stats", func(w http.ResponseWriter, r *http.Request) {
		if deps.ReplayStatsStore == nil {
			http.Error(w, "replay stats store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			submissionID, err := parsePathID(r)
			if err != nil {
				http.Error(w, "invalid submission id", http.StatusBadRequest)
				return
			}
			lines, err := deps.ReplayStatsStore.ListStatsBySubmission(r.Context(), submissionID)
			if err != nil {
				http.Error(w, "failed to list stats", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, lines)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/players/{id}/career-stats", func(w http.ResponseWriter, r *http.Request) {
		if deps.ReplayStatsStore == nil {
			http.Error(w, "replay stats store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			playerID, err := parsePathID(r)
			if err != nil {
				http.Error(w, "invalid player id", http.StatusBadRequest)
				return
			}
			stats, err := deps.ReplayStatsStore.GetPlayerCareerStats(r.Context(), playerID)
			if err != nil {
				http.Error(w, "failed to get player career stats", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, stats)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/exceptions/report", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input hierarchy.ReportExceptionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			ticket, err := deps.HierarchyStore.ReportException(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, ticket)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/operator-inbox", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			tickets, err := deps.HierarchyStore.ListOperatorInbox(r.Context())
			if err != nil {
				http.Error(w, "failed to list operator inbox", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, tickets)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/operator-inbox/triage", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input hierarchy.TriageExceptionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			ticket, err := deps.HierarchyStore.TriageException(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, ticket)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/operator-inbox/resolve", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input hierarchy.ResolveExceptionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			ticket, err := deps.HierarchyStore.ResolveException(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, ticket)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/exception-actions", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			actions, err := deps.HierarchyStore.ListExceptionActions(r.Context())
			if err != nil {
				http.Error(w, "failed to list exception actions", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, actions)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/exception-metrics", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			metrics, err := deps.HierarchyStore.GetExceptionMetrics(r.Context())
			if err != nil {
				http.Error(w, "failed to get exception metrics", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, metrics)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/exception-automations/scheduling", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input hierarchy.EvaluateSchedulingExceptionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			result, err := deps.HierarchyStore.EvaluateSchedulingException(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, result)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/exception-automations/no-show", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input hierarchy.EvaluateNoShowExceptionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			result, err := deps.HierarchyStore.EvaluateNoShowException(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, result)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/exception-automations/replay-dispute", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input hierarchy.EvaluateReplayDisputeExceptionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			result, err := deps.HierarchyStore.EvaluateReplayDisputeException(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, result)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/seasons", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			seasons, err := deps.HierarchyStore.ListSeasons(r.Context())
			if err != nil {
				http.Error(w, "failed to list seasons", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, seasons)
		case http.MethodPost:
			var input hierarchy.CreateSeasonInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			season, err := deps.HierarchyStore.CreateSeason(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, season)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/schedule-groups", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			groups, err := deps.HierarchyStore.ListScheduleGroups(r.Context())
			if err != nil {
				http.Error(w, "failed to list schedule groups", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, groups)
		case http.MethodPost:
			var input hierarchy.CreateScheduleGroupInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			group, err := deps.HierarchyStore.CreateScheduleGroup(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, group)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/fixtures", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			fixtures, err := deps.HierarchyStore.ListFixtures(r.Context())
			if err != nil {
				http.Error(w, "failed to list fixtures", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, fixtures)
		case http.MethodPost:
			var input hierarchy.CreateFixtureInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			fixture, err := deps.HierarchyStore.CreateFixture(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, fixture)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/matches", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			matches, err := deps.HierarchyStore.ListMatches(r.Context())
			if err != nil {
				http.Error(w, "failed to list matches", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, matches)
		case http.MethodPost:
			var input hierarchy.CreateMatchInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			match, err := deps.HierarchyStore.CreateMatch(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, match)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// --- Single-resource lookups (Theme 7.1) ---

	mux.HandleFunc("GET /v1/fixtures/{id}", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		id, err := parsePathID(r)
		if err != nil {
			http.Error(w, "invalid fixture id", http.StatusBadRequest)
			return
		}
		fixture, err := deps.HierarchyStore.GetFixture(r.Context(), id)
		if err != nil {
			handleHierarchyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, fixture)
	})

	mux.HandleFunc("GET /v1/scrims/{id}", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		id, err := parsePathID(r)
		if err != nil {
			http.Error(w, "invalid scrim id", http.StatusBadRequest)
			return
		}
		scrim, err := deps.HierarchyStore.GetScrim(r.Context(), id)
		if err != nil {
			handleHierarchyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, scrim)
	})

	mux.HandleFunc("GET /v1/result-submissions/{id}", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		id, err := parsePathID(r)
		if err != nil {
			http.Error(w, "invalid submission id", http.StatusBadRequest)
			return
		}
		submission, err := deps.HierarchyStore.GetResultSubmission(r.Context(), id)
		if err != nil {
			handleHierarchyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, submission)
	})

	// --- Submission reset (Theme 7.2) ---

	mux.HandleFunc("POST /v1/result-submissions/{id}/reset", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		if !checkPermission(w, r, sessionCookieName, sessionSecret, tokenValidator, deps.APITokenStore, evaluator, authz.ResourceResultSubmission, authz.ActionReset) {
			return
		}
		id, err := parsePathID(r)
		if err != nil {
			http.Error(w, "invalid submission id", http.StatusBadRequest)
			return
		}
		var body struct {
			Actor string `json:"actor"`
		}
		if err := decodeJSON(r, &body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if body.Actor == "" {
			if principal, ok := readSessionPrincipal(r, sessionCookieName, sessionSecret); ok {
				body.Actor = principal.Subject
			}
		}
		submission, err := deps.HierarchyStore.ResetResultSubmission(r.Context(), hierarchy.ResetResultSubmissionInput{
			SubmissionID: id,
			Actor:        body.Actor,
		})
		if err != nil {
			handleHierarchyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, submission)
	})

	// --- Player activate / deactivate (Theme 7.3) ---

	mux.HandleFunc("POST /v1/players/{id}/activate", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		if !checkPermission(w, r, sessionCookieName, sessionSecret, tokenValidator, deps.APITokenStore, evaluator, authz.ResourcePlayer, authz.ActionUpdate) {
			return
		}
		id, err := parsePathID(r)
		if err != nil {
			http.Error(w, "invalid player id", http.StatusBadRequest)
			return
		}
		player, err := deps.HierarchyStore.SetPlayerActive(r.Context(), hierarchy.SetPlayerActiveInput{
			PlayerID: id,
			IsActive: true,
		})
		if err != nil {
			handleHierarchyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, player)
	})

	mux.HandleFunc("POST /v1/players/{id}/deactivate", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		if !checkPermission(w, r, sessionCookieName, sessionSecret, tokenValidator, deps.APITokenStore, evaluator, authz.ResourcePlayer, authz.ActionUpdate) {
			return
		}
		id, err := parsePathID(r)
		if err != nil {
			http.Error(w, "invalid player id", http.StatusBadRequest)
			return
		}
		player, err := deps.HierarchyStore.SetPlayerActive(r.Context(), hierarchy.SetPlayerActiveInput{
			PlayerID: id,
			IsActive: false,
		})
		if err != nil {
			handleHierarchyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, player)
	})

	// --- Skill group transitions (Theme 5C) ---

	mux.HandleFunc("GET /v1/players/{id}/skill-group-transitions", func(w http.ResponseWriter, r *http.Request) {
		if deps.SkillGroupStore == nil {
			http.Error(w, "skill group store unavailable", http.StatusServiceUnavailable)
			return
		}
		playerID, err := parsePathID(r)
		if err != nil {
			http.Error(w, "invalid player id", http.StatusBadRequest)
			return
		}
		transitions, err := deps.SkillGroupStore.ListTransitionsByPlayerID(r.Context(), playerID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, transitions)
	})

	// --- Skill groups (Theme 5A) ---

	mux.HandleFunc("/v1/leagues/{id}/skill-groups", func(w http.ResponseWriter, r *http.Request) {
		if deps.SkillGroupStore == nil {
			http.Error(w, "skill group store unavailable", http.StatusServiceUnavailable)
			return
		}
		leagueID, err := parsePathID(r)
		if err != nil {
			http.Error(w, "invalid league id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			groups, err := deps.SkillGroupStore.ListSkillGroups(r.Context(), leagueID)
			if err != nil {
				http.Error(w, "failed to list skill groups", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, groups)
		case http.MethodPost:
			var input skillgroup.CreateSkillGroupInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			input.LeagueID = leagueID
			sg, err := deps.SkillGroupStore.CreateSkillGroup(r.Context(), input)
			if err != nil {
				handleSkillGroupError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, sg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/skill-groups/{id}", func(w http.ResponseWriter, r *http.Request) {
		if deps.SkillGroupStore == nil {
			http.Error(w, "skill group store unavailable", http.StatusServiceUnavailable)
			return
		}
		id, err := parsePathID(r)
		if err != nil {
			http.Error(w, "invalid skill group id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			sg, err := deps.SkillGroupStore.GetSkillGroup(r.Context(), id)
			if err != nil {
				handleSkillGroupError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, sg)
		case http.MethodPatch:
			var input skillgroup.UpdateSkillGroupInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			input.SkillGroupID = id
			sg, err := deps.SkillGroupStore.UpdateSkillGroup(r.Context(), input)
			if err != nil {
				handleSkillGroupError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, sg)
		case http.MethodDelete:
			inactive := false
			sg, err := deps.SkillGroupStore.UpdateSkillGroup(r.Context(), skillgroup.UpdateSkillGroupInput{
				SkillGroupID: id,
				IsActive:     &inactive,
			})
			if err != nil {
				handleSkillGroupError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, sg)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// --- Organization config (Theme 6) ---

	mux.HandleFunc("/v1/leagues/{id}/config", func(w http.ResponseWriter, r *http.Request) {
		if deps.OrgConfigStore == nil {
			http.Error(w, "org config store unavailable", http.StatusServiceUnavailable)
			return
		}
		leagueID, err := parsePathID(r)
		if err != nil {
			http.Error(w, "invalid league id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			configs, err := deps.OrgConfigStore.ListConfigs(r.Context(), leagueID)
			if err != nil {
				http.Error(w, "failed to read config", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, configs)
		case http.MethodPatch:
			var updates map[string]string
			if err := decodeJSON(r, &updates); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			var updatedBy *int64
			if principal, ok := readSessionPrincipal(r, sessionCookieName, sessionSecret); ok {
				playerID, parseErr := strconv.ParseInt(principal.Subject, 10, 64)
				if parseErr == nil {
					updatedBy = &playerID
				}
			}
			for key, value := range updates {
				if err := deps.OrgConfigStore.SetConfig(r.Context(), leagueID, key, value, updatedBy); err != nil {
					http.Error(w, "failed to update config key: "+key, http.StatusInternalServerError)
					return
				}
			}
			configs, err := deps.OrgConfigStore.ListConfigs(r.Context(), leagueID)
			if err != nil {
				http.Error(w, "failed to read updated config", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, configs)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// --- Scrim check-in (Theme 3.2) ---

	mux.HandleFunc("POST /v1/scrims/{id}/check-in", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		scrimID, err := parsePathID(r)
		if err != nil {
			http.Error(w, "invalid scrim id", http.StatusBadRequest)
			return
		}
		var body struct {
			TeamID int64 `json:"teamId"`
		}
		if err := decodeJSON(r, &body); err != nil || body.TeamID == 0 {
			http.Error(w, "teamId is required", http.StatusBadRequest)
			return
		}
		scrim, err := deps.HierarchyStore.CheckInScrim(r.Context(), hierarchy.CheckInScrimInput{
			ScrimID: scrimID,
			TeamID:  body.TeamID,
		})
		if err != nil {
			handleHierarchyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, scrim)
	})

	// --- Scrim metrics (Theme 3.4) ---

	mux.HandleFunc("GET /v1/scrim-metrics", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		metrics, err := deps.HierarchyStore.GetScrimMetrics(r.Context())
		if err != nil {
			http.Error(w, "failed to get scrim metrics", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, metrics)
	})

	// --- Me-scoped endpoints (Theme 2) ---

	mux.HandleFunc("GET /v1/me", func(w http.ResponseWriter, r *http.Request) {
		principal, ok := readSessionPrincipal(r, sessionCookieName, sessionSecret)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"subject":     principal.Subject,
			"displayName": principal.DisplayName,
			"roles":       principal.Roles,
		})
	})

	mux.HandleFunc("GET /v1/me/eligibility", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		principal, ok := readSessionPrincipal(r, sessionCookieName, sessionSecret)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		status, err := deps.HierarchyStore.GetEligibilityStatus(r.Context(), principal.Subject)
		if err != nil {
			handleHierarchyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, status)
	})

	mux.HandleFunc("GET /v1/me/roster-membership", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		playerID, err := parseIntQueryParam(r, "player_id")
		if err != nil {
			http.Error(w, "player_id query parameter required", http.StatusBadRequest)
			return
		}
		rm, err := deps.HierarchyStore.GetActiveRosterMembershipByPlayerID(r.Context(), playerID)
		if err != nil {
			http.Error(w, "failed to get roster membership", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, rm)
	})

	mux.HandleFunc("GET /v1/me/queue-bans", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		playerID, err := parseIntQueryParam(r, "player_id")
		if err != nil {
			http.Error(w, "player_id query parameter required", http.StatusBadRequest)
			return
		}
		bans, err := deps.HierarchyStore.ListActiveQueueBansByPlayerID(r.Context(), playerID)
		if err != nil {
			http.Error(w, "failed to get queue bans", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, bans)
	})

	mux.HandleFunc("GET /v1/me/queue-entry", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		playerID, err := parseIntQueryParam(r, "player_id")
		if err != nil {
			http.Error(w, "player_id query parameter required", http.StatusBadRequest)
			return
		}
		entry, err := deps.HierarchyStore.GetActiveQueueEntryByPlayerID(r.Context(), playerID)
		if err != nil {
			http.Error(w, "failed to get queue entry", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, entry)
	})

	mux.HandleFunc("GET /v1/me/scrim", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		playerID, err := parseIntQueryParam(r, "player_id")
		if err != nil {
			http.Error(w, "player_id query parameter required", http.StatusBadRequest)
			return
		}
		scrim, err := deps.HierarchyStore.GetActiveScrimByPlayerID(r.Context(), playerID)
		if err != nil {
			http.Error(w, "failed to get scrim", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, scrim)
	})

	mux.HandleFunc("GET /v1/me/result-submissions", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}
		playerID, err := parseIntQueryParam(r, "player_id")
		if err != nil {
			http.Error(w, "player_id query parameter required", http.StatusBadRequest)
			return
		}
		rm, err := deps.HierarchyStore.GetActiveRosterMembershipByPlayerID(r.Context(), playerID)
		if err != nil {
			http.Error(w, "failed to resolve roster membership", http.StatusInternalServerError)
			return
		}
		if rm == nil {
			writeJSON(w, http.StatusOK, []hierarchy.ResultSubmission{})
			return
		}
		input := hierarchy.ListResultSubmissionsInput{TeamID: &rm.TeamID}
		if state := strings.TrimSpace(r.URL.Query().Get("state")); state != "" {
			input.State = &state
		}
		submissions, err := deps.HierarchyStore.ListResultSubmissionsFiltered(r.Context(), input)
		if err != nil {
			http.Error(w, "failed to list result submissions", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, submissions)
	})

	// --- Notifications (Theme 2.3) ---

	mux.HandleFunc("GET /v1/me/notifications", func(w http.ResponseWriter, r *http.Request) {
		if deps.NotificationsStore == nil {
			http.Error(w, "notifications store unavailable", http.StatusServiceUnavailable)
			return
		}
		playerID, err := parseIntQueryParam(r, "player_id")
		if err != nil {
			http.Error(w, "player_id query parameter required", http.StatusBadRequest)
			return
		}
		unreadOnly := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("unread")), "true")
		notifs, err := deps.NotificationsStore.ListNotifications(r.Context(), playerID, unreadOnly)
		if err != nil {
			http.Error(w, "failed to list notifications", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, notifs)
	})

	mux.HandleFunc("POST /v1/me/notifications/mark-read", func(w http.ResponseWriter, r *http.Request) {
		if deps.NotificationsStore == nil {
			http.Error(w, "notifications store unavailable", http.StatusServiceUnavailable)
			return
		}
		playerID, err := parseIntQueryParam(r, "player_id")
		if err != nil {
			http.Error(w, "player_id query parameter required", http.StatusBadRequest)
			return
		}
		var body struct {
			IDs     []int64 `json:"ids"`
			MarkAll bool    `json:"all"`
		}
		if err := decodeJSON(r, &body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if err := deps.NotificationsStore.MarkRead(r.Context(), notifications.MarkReadInput{
			PlayerID: playerID,
			IDs:      body.IDs,
			MarkAll:  body.MarkAll,
		}); err != nil {
			http.Error(w, "failed to mark notifications as read", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	// --- API Tokens (Theme 1.4) ---

	mux.HandleFunc("POST /v1/api-tokens", func(w http.ResponseWriter, r *http.Request) {
		if deps.APITokenStore == nil {
			http.Error(w, "api token store unavailable", http.StatusServiceUnavailable)
			return
		}
		principal, ok := readPrincipal(r, sessionCookieName, sessionSecret, deps.APITokenStore)
		if !ok {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		var body struct {
			Name   string   `json:"name"`
			Scopes []string `json:"scopes"`
		}
		if err := decodeJSON(r, &body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if strings.TrimSpace(body.Name) == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		input := apitoken.CreateAPITokenInput{
			Name:    body.Name,
			Subject: principal.Subject,
			Scopes:  body.Scopes,
		}
		token, plainText, err := deps.APITokenStore.CreateAPIToken(r.Context(), input)
		if err != nil {
			http.Error(w, "failed to create api token", http.StatusInternalServerError)
			return
		}
		type createResponse struct {
			apitoken.APIToken
			PlainTextToken string `json:"plainTextToken"`
		}
		writeJSON(w, http.StatusCreated, createResponse{APIToken: token, PlainTextToken: plainText})
	})

	mux.HandleFunc("GET /v1/api-tokens", func(w http.ResponseWriter, r *http.Request) {
		if deps.APITokenStore == nil {
			http.Error(w, "api token store unavailable", http.StatusServiceUnavailable)
			return
		}
		principal, ok := readPrincipal(r, sessionCookieName, sessionSecret, deps.APITokenStore)
		if !ok {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		tokens, err := deps.APITokenStore.ListAPITokens(r.Context(), principal.Subject)
		if err != nil {
			http.Error(w, "failed to list api tokens", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, tokens)
	})

	mux.HandleFunc("DELETE /v1/api-tokens/{id}", func(w http.ResponseWriter, r *http.Request) {
		if deps.APITokenStore == nil {
			http.Error(w, "api token store unavailable", http.StatusServiceUnavailable)
			return
		}
		principal, ok := readPrincipal(r, sessionCookieName, sessionSecret, deps.APITokenStore)
		if !ok {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		idStr := r.PathValue("id")
		tokenID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid token id", http.StatusBadRequest)
			return
		}
		token, err := deps.APITokenStore.RevokeAPIToken(r.Context(), apitoken.RevokeAPITokenInput{
			TokenID: tokenID,
			Subject: principal.Subject,
		})
		if err != nil {
			if errors.Is(err, apitoken.ErrNotFound) {
				http.Error(w, "token not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to revoke api token", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, token)
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return &Server{logger: logger, http: srv}
}

func (s *Server) Start() error {
	s.logger.Info("http server starting", "addr", s.http.Addr)
	return s.http.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("http server shutting down")
	return s.http.Shutdown(ctx)
}

func (s *Server) Handler() http.Handler {
	return s.http.Handler
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func decodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func handleHierarchyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, hierarchy.ErrInvalidInput):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, hierarchy.ErrConflict), errors.Is(err, hierarchy.ErrDependency):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func handleSkillGroupError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, skillgroup.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, skillgroup.ErrInvalidInput):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, skillgroup.ErrConflict), errors.Is(err, skillgroup.ErrDependency):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func parsePathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func parseIntQueryParam(r *http.Request, name string) (int64, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return 0, errors.New("missing")
	}
	return strconv.ParseInt(raw, 10, 64)
}

func parseSessionTTL(raw string) time.Duration {
	parsed, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil || parsed <= 0 {
		return 12 * time.Hour
	}
	return parsed
}

func readSessionPrincipal(r *http.Request, cookieName, secret string) (auth.SessionPrincipal, bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return auth.SessionPrincipal{}, false
	}
	principal, err := auth.VerifySessionToken(cookie.Value, secret, time.Now().UTC())
	if err != nil {
		return auth.SessionPrincipal{}, false
	}
	return principal, true
}

// readPrincipal returns an auth.SessionPrincipal for the request. It first
// checks the session cookie; if none is present it falls back to an
// Authorization: Bearer token validated via the apitoken store. Session cookie
// auth takes priority.
func readPrincipal(r *http.Request, cookieName, secret string, tokenStore apitoken.Store) (auth.SessionPrincipal, bool) {
	if p, ok := readSessionPrincipal(r, cookieName, secret); ok {
		return p, true
	}
	if tokenStore == nil {
		return auth.SessionPrincipal{}, false
	}
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return auth.SessionPrincipal{}, false
	}
	rawToken := strings.TrimPrefix(authHeader, "Bearer ")
	rawToken = strings.TrimSpace(rawToken)
	if rawToken == "" {
		return auth.SessionPrincipal{}, false
	}
	result, err := tokenStore.ValidateAPIToken(r.Context(), rawToken)
	if err != nil {
		return auth.SessionPrincipal{}, false
	}
	return auth.SessionPrincipal{
		Subject:     result.Subject,
		DisplayName: result.Subject,
		Roles:       result.Scopes,
		ExpiresAt:   time.Time{},
	}, true
}

func buildAuthCallbackQuery(values url.Values, webBaseURL string) string {
	subject := strings.TrimSpace(values.Get("subject"))
	if subject == "" {
		subject = "local-player"
	}
	displayName := strings.TrimSpace(values.Get("displayName"))
	roles := strings.TrimSpace(values.Get("roles"))
	if roles == "" {
		roles = "player"
	}
	redirect := normalizeRedirectURL(values.Get("redirect"), webBaseURL)
	query := url.Values{}
	query.Set("subject", subject)
	query.Set("displayName", firstNonEmpty(displayName, subject))
	query.Set("roles", roles)
	query.Set("redirect", redirect)
	return query.Encode()
}

func normalizeRedirectURL(raw, webBaseURL string) string {
	defaultRedirect := strings.TrimRight(webBaseURL, "/") + "/app"
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return defaultRedirect
	}
	if strings.HasPrefix(trimmed, "/") {
		return strings.TrimRight(webBaseURL, "/") + trimmed
	}

	targetURL, err := url.Parse(trimmed)
	if err != nil || targetURL.Scheme == "" || targetURL.Host == "" {
		return defaultRedirect
	}

	allowedURL, err := url.Parse(webBaseURL)
	if err != nil || !strings.EqualFold(allowedURL.Host, targetURL.Host) {
		return defaultRedirect
	}
	return targetURL.String()
}

func parseRoleList(raw string) []string {
	roles := make([]string, 0)
	for _, role := range strings.Split(raw, ",") {
		trimmed := strings.TrimSpace(role)
		if trimmed != "" {
			roles = append(roles, trimmed)
		}
	}
	if len(roles) == 0 {
		return []string{"player"}
	}
	return roles
}

// startPopTimeoutWatcher launches a goroutine that cancels a popped scrim if it hasn't
// fully checked in within the given timeout. Default ban duration is 60 minutes.
func startPopTimeoutWatcher(scrimID int64, timeout time.Duration, store hierarchy.Store, logger *slog.Logger) {
	time.AfterFunc(timeout, func() {
		ctx := context.Background()
		if err := store.ExecutePopTimeout(ctx, hierarchy.ExecutePopTimeoutInput{
			ScrimID:                 scrimID,
			QueueBanDurationMinutes: 60,
		}); err != nil {
			logger.Error("pop timeout execution failed", "scrimId", scrimID, "error", err)
		}
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// checkPermission reads the requesting principal and checks whether they hold
// the required resource+action permission. Auth is resolved in order:
//  1. Session cookie
//  2. Bearer token validated by the TokenValidator (local/JWT format)
//  3. Bearer token validated by the APITokenStore (spr_ machine tokens)
//
// It writes an appropriate HTTP error and returns false if the check fails.
// Callers should return immediately when this returns false.
func checkPermission(
	w http.ResponseWriter,
	r *http.Request,
	cookieName, secret string,
	tokenValidator auth.TokenValidator,
	tokenStore apitoken.Store,
	ev authz.Evaluator,
	resource, action string,
) bool {
	var roles []string

	if sp, ok := readSessionPrincipal(r, cookieName, secret); ok {
		roles = sp.Roles
	} else if tokenValidator != nil {
		authHeader := r.Header.Get("Authorization")
		if principal, err := tokenValidator.Validate(authHeader); err == nil && !principal.Anonymous {
			roles = principal.Roles
		} else if tokenStore != nil {
			sp, ok := readAPIBearerPrincipal(r, tokenStore)
			if !ok {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return false
			}
			roles = sp.Roles
		} else {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return false
		}
	} else if tokenStore != nil {
		sp, ok := readAPIBearerPrincipal(r, tokenStore)
		if !ok {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return false
		}
		roles = sp.Roles
	} else {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return false
	}

	if ev == nil {
		// No evaluator configured — allow by default (development mode).
		return true
	}
	if !ev.AllowedInContext(roles, nil, resource, action, authz.GlobalContext()) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

// readAPIBearerPrincipal extracts a principal from an Authorization: Bearer
// header using the APITokenStore (for spr_* machine tokens).
func readAPIBearerPrincipal(r *http.Request, tokenStore apitoken.Store) (auth.SessionPrincipal, bool) {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return auth.SessionPrincipal{}, false
	}
	rawToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if rawToken == "" {
		return auth.SessionPrincipal{}, false
	}
	result, err := tokenStore.ValidateAPIToken(r.Context(), rawToken)
	if err != nil {
		return auth.SessionPrincipal{}, false
	}
	return auth.SessionPrincipal{
		Subject:     result.Subject,
		DisplayName: result.Subject,
		Roles:       result.Scopes,
		ExpiresAt:   time.Time{},
	}, true
}

const oauthStateCookieName = "oauth_state"

// generateOAuthState returns a cryptographically random hex string suitable
// for use as an OAuth2 state parameter.
func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// setOAuthStateCookie writes a short-lived HttpOnly cookie containing the
// OAuth state value. The cookie expires in 10 minutes.
func setOAuthStateCookie(w http.ResponseWriter, state string) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// validateOAuthStateCookie checks that the provided state value matches the
// oauth_state cookie on the request. It returns true only when the cookie is
// present and the values match.
func validateOAuthStateCookie(r *http.Request, providedState string) bool {
	if strings.TrimSpace(providedState) == "" {
		return false
	}
	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		return false
	}
	return cookie.Value == providedState
}
