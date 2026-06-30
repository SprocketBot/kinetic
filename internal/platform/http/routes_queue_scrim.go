package http

import (
	"github.com/sprocketbot/sprocket-v3/internal/domain/authz"
	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
	"net/http"
	"time"
)

func (r routeRegistrar) registerQueueAndScrimRoutes(mux *http.ServeMux) {
	logger := r.logger
	deps := r.deps
	tokenValidator := r.tokenValidator
	evaluator := r.evaluator
	sessionCookieName := r.sessionCookieName
	sessionSecret := r.sessionSecret

	mux.HandleFunc("/v1/queue-entries", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			entries, err := deps.QueueStore.ListActiveQueueEntries(r.Context())
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
			entry, err := deps.QueueStore.EnqueueTeam(r.Context(), input)
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
			entry, err := deps.QueueStore.LeaveQueue(r.Context(), input)
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
			entry, err := deps.QueueStore.AdvanceQueueEntryStage(r.Context(), input)
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
			bans, err := deps.QueueStore.ListQueueBans(r.Context())
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
			ban, err := deps.QueueStore.BanPlayerFromQueue(r.Context(), input)
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
			ban, err := deps.QueueStore.UnbanPlayerFromQueue(r.Context(), input)
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
			scrims, err := deps.ScrimStore.ListScrims(r.Context())
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
			scrim, err := deps.ScrimStore.CreateScrim(r.Context(), input)
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
			scrim, err := deps.ScrimStore.UpdateScrimState(r.Context(), input)
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
			scrim, err := deps.ScrimStore.PromoteQueueToScrim(r.Context(), input)
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
			result, err := deps.ScrimStore.ProcessQueuePromotions(r.Context(), input)
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
			runs, err := deps.ScrimStore.ListPromotionProcessingRuns(r.Context())
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
			ratings, err := deps.RatingStore.ListPlayerRatings(r.Context())
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
			rating, err := deps.RatingStore.AdjustPlayerRating(r.Context(), input)
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
			adjustments, err := deps.RatingStore.ListRatingAdjustments(r.Context())
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
			decisions, err := deps.RatingStore.ListMatchmakingDecisions(r.Context())
			if err != nil {
				http.Error(w, "failed to list matchmaking decisions", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, decisions)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
