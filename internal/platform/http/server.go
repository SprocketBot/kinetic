package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sprocketbot/sprocket-v3/internal/domain/authz"
	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
	"github.com/sprocketbot/sprocket-v3/internal/platform/auth"
	"github.com/sprocketbot/sprocket-v3/internal/platform/config"
)

type Server struct {
	logger *slog.Logger
	http   *http.Server
}

type Dependencies struct {
	TokenValidator auth.TokenValidator
	Evaluator      authz.Evaluator
	HierarchyStore hierarchy.Store
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

	mux.HandleFunc("/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			callbackURL := "/v1/auth/callback?" + buildAuthCallbackQuery(r.URL.Query(), webBaseURL)
			http.Redirect(w, r, callbackURL, http.StatusFound)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			query := r.URL.Query()
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
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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
			)
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
