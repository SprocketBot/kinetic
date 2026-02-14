package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
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
