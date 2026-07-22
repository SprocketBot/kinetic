package http

import (
	"net/http"

	"github.com/kineticbot/kinetic-v3/internal/domain/authz"
	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
)

func (r routeRegistrar) registerIdentityRoutes(mux *http.ServeMux) {
	deps := r.deps

	mux.HandleFunc("/v1/games", func(w http.ResponseWriter, req *http.Request) {
		if deps.IdentityStore == nil {
			http.Error(w, "identity store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch req.Method {
		case http.MethodGet:
			games, err := deps.IdentityStore.ListGames(req.Context())
			if err != nil {
				http.Error(w, "failed to list games", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, games)
		case http.MethodPost:
			if !checkPermission(w, req, r.sessionCookieName, r.sessionSecret, r.tokenValidator, deps.APITokenStore, r.evaluator, authz.ResourceGame, authz.ActionCreate) {
				return
			}
			var input hierarchy.CreateGameInput
			if err := decodeJSON(req, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			game, err := deps.IdentityStore.CreateGame(req.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, game)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/v1/me/players", func(w http.ResponseWriter, req *http.Request) {
		if deps.IdentityStore == nil {
			http.Error(w, "identity store unavailable", http.StatusServiceUnavailable)
			return
		}
		principal, ok := readRequestPrincipal(req, r.sessionCookieName, r.sessionSecret, r.tokenValidator, deps.APITokenStore)
		if !ok {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		user, err := deps.IdentityStore.UpsertUser(req.Context(), hierarchy.UpsertUserInput{
			Subject: principal.Subject, DisplayName: principal.DisplayName,
		})
		if err != nil {
			http.Error(w, "failed to resolve user", http.StatusInternalServerError)
			return
		}

		switch req.Method {
		case http.MethodGet:
			players, err := deps.IdentityStore.ListUserPlayers(req.Context(), user.ID)
			if err != nil {
				http.Error(w, "failed to list user players", http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, players)
		case http.MethodPost:
			var input hierarchy.CreateUserPlayerInput
			if err := decodeJSON(req, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			input.UserID = user.ID
			player, err := deps.IdentityStore.CreateUserPlayer(req.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusCreated, player)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
