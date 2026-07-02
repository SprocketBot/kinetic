package http

import (
	"github.com/kineticbot/kinetic-v3/internal/domain/authz"
	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
	"github.com/kineticbot/kinetic-v3/internal/domain/skillgroup"
	"net/http"
	"strconv"
)

func (r routeRegistrar) registerAdminMutationRoutes(mux *http.ServeMux) {
	deps := r.deps
	tokenValidator := r.tokenValidator
	evaluator := r.evaluator
	sessionCookieName := r.sessionCookieName
	sessionSecret := r.sessionSecret

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
		scrim, err := deps.ScrimStore.GetScrim(r.Context(), id)
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
		submission, err := deps.ResultStore.GetResultSubmission(r.Context(), id)
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
		submission, err := deps.ResultStore.ResetResultSubmission(r.Context(), hierarchy.ResetResultSubmissionInput{
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
		player, err := deps.PlayerStore.SetPlayerActive(r.Context(), hierarchy.SetPlayerActiveInput{
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
		player, err := deps.PlayerStore.SetPlayerActive(r.Context(), hierarchy.SetPlayerActiveInput{
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
		scrim, err := deps.ScrimStore.CheckInScrim(r.Context(), hierarchy.CheckInScrimInput{
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
		metrics, err := deps.ScrimStore.GetScrimMetrics(r.Context())
		if err != nil {
			http.Error(w, "failed to get scrim metrics", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, metrics)
	})
}
