package http

import (
	"errors"
	"github.com/kineticbot/kinetic-v3/internal/domain/apitoken"
	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
	"github.com/kineticbot/kinetic-v3/internal/domain/notifications"
	"net/http"
	"strconv"
	"strings"
)

func (r routeRegistrar) registerSelfServiceAndTokenRoutes(mux *http.ServeMux) {
	deps := r.deps
	sessionCookieName := r.sessionCookieName
	sessionSecret := r.sessionSecret

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
		status, err := deps.EligibilityStore.GetEligibilityStatus(r.Context(), principal.Subject)
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
		rm, err := deps.RosterStore.GetActiveRosterMembershipByPlayerID(r.Context(), playerID)
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
		bans, err := deps.QueueStore.ListActiveQueueBansByPlayerID(r.Context(), playerID)
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
		entry, err := deps.QueueStore.GetActiveQueueEntryByPlayerID(r.Context(), playerID)
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
		scrim, err := deps.ScrimStore.GetActiveScrimByPlayerID(r.Context(), playerID)
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
		rm, err := deps.RosterStore.GetActiveRosterMembershipByPlayerID(r.Context(), playerID)
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
		submissions, err := deps.ResultStore.ListResultSubmissionsFiltered(r.Context(), input)
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
}
