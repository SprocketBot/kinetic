package http

import (
	"github.com/kineticbot/kinetic-v3/internal/domain/authz"
	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
	"github.com/kineticbot/kinetic-v3/internal/platform/auth"
	"net/http"
	"strconv"
	"strings"
)

func (r routeRegistrar) registerHierarchyRoutes(mux *http.ServeMux) {
	deps := r.deps
	tokenValidator := r.tokenValidator
	evaluator := r.evaluator
	sessionCookieName := r.sessionCookieName
	sessionSecret := r.sessionSecret

	mux.HandleFunc("/v1/leagues", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			leagues, err := deps.LeagueStore.ListLeagues(r.Context())
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
			league, err := deps.LeagueStore.CreateLeague(r.Context(), input)
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
			franchises, err := deps.LeagueStore.ListFranchises(r.Context())
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
			franchise, err := deps.LeagueStore.CreateFranchise(r.Context(), input)
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
			clubs, err := deps.LeagueStore.ListClubs(r.Context())
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
			club, err := deps.LeagueStore.CreateClub(r.Context(), input)
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
			teams, err := deps.LeagueStore.ListTeams(r.Context())
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
			team, err := deps.LeagueStore.CreateTeam(r.Context(), input)
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
			players, err := deps.PlayerStore.ListPlayers(r.Context())
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
			player, err := deps.PlayerStore.CreatePlayer(r.Context(), input)
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
			memberships, err := deps.RosterStore.ListRosterMemberships(r.Context())
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
			scope, err := deps.RoleStore.ResolveTeamScope(r.Context(), input.TeamID)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			if !authorizeScopedAction(w, r, deps, tokenValidator, evaluator, sessionCookieName, sessionSecret, scope, authz.ResourceRosterMembership, authz.ActionCreate) {
				return
			}
			membership, err := deps.RosterStore.CreateRosterMembership(r.Context(), input)
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
			assignments, err := deps.RoleStore.ListRoleAssignments(r.Context())
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
			scope, err := deps.RoleStore.ResolveRoleScope(r.Context(), input.Role, input.FranchiseID, input.ClubID, input.TeamID)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			actorPlayerID, ok := authorizeScopedRoleMutation(w, r, deps, tokenValidator, evaluator, sessionCookieName, sessionSecret, scope, authz.ActionCreate)
			if !ok {
				return
			}
			input.ActorPlayerID = actorPlayerID
			assignment, err := deps.RoleStore.AssignRole(r.Context(), input)
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
			scope, err := deps.RoleStore.ResolveAssignmentScope(r.Context(), input.AssignmentID)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			actorPlayerID, ok := authorizeScopedRoleMutation(w, r, deps, tokenValidator, evaluator, sessionCookieName, sessionSecret, scope, authz.ActionRevoke)
			if !ok {
				return
			}
			input.ActorPlayerID = actorPlayerID
			assignment, err := deps.RoleStore.RevokeRole(r.Context(), input)
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
			queues, err := deps.QueueStore.ListQueues(r.Context())
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
			queue, err := deps.QueueStore.CreateQueue(r.Context(), input)
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
			if playerIDRaw := strings.TrimSpace(r.URL.Query().Get("playerId")); playerIDRaw != "" {
				playerID, err := strconv.ParseInt(playerIDRaw, 10, 64)
				if err != nil || playerID <= 0 {
					http.Error(w, "playerId must be a positive integer", http.StatusBadRequest)
					return
				}
				principal, ok := readRequestPrincipal(r, sessionCookieName, sessionSecret, tokenValidator, deps.APITokenStore)
				if !ok {
					http.Error(w, "authentication required", http.StatusUnauthorized)
					return
				}
				user, err := deps.IdentityStore.UpsertUser(r.Context(), hierarchy.UpsertUserInput{Subject: principal.Subject, DisplayName: principal.DisplayName})
				if err != nil {
					http.Error(w, "failed to resolve user", http.StatusInternalServerError)
					return
				}
				owns, err := deps.IdentityStore.UserOwnsPlayer(r.Context(), user.ID, playerID)
				if err != nil {
					http.Error(w, "failed to verify player ownership", http.StatusInternalServerError)
					return
				}
				if !owns {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
				links, err := deps.PlatformStore.ListPlatformAccountLinksByPlayerID(r.Context(), playerID)
				if err != nil {
					http.Error(w, "failed to list platform account links", http.StatusInternalServerError)
					return
				}
				writeJSON(w, http.StatusOK, links)
				return
			}
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
			links, err := deps.PlatformStore.ListPlatformAccountLinks(r.Context(), subject)
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
			if input.PlayerID != nil {
				principal, ok := readRequestPrincipal(r, sessionCookieName, sessionSecret, tokenValidator, deps.APITokenStore)
				if !ok {
					http.Error(w, "authentication required", http.StatusUnauthorized)
					return
				}
				user, err := deps.IdentityStore.UpsertUser(r.Context(), hierarchy.UpsertUserInput{Subject: principal.Subject, DisplayName: principal.DisplayName})
				if err != nil {
					http.Error(w, "failed to resolve user", http.StatusInternalServerError)
					return
				}
				owns, err := deps.IdentityStore.UserOwnsPlayer(r.Context(), user.ID, *input.PlayerID)
				if err != nil {
					http.Error(w, "failed to verify player ownership", http.StatusInternalServerError)
					return
				}
				if !owns {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
				input.Subject = principal.Subject
			}
			link, err := deps.PlatformStore.LinkPlatformAccount(r.Context(), input)
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
			if input.PlayerID != nil {
				principal, ok := readRequestPrincipal(r, sessionCookieName, sessionSecret, tokenValidator, deps.APITokenStore)
				if !ok {
					http.Error(w, "authentication required", http.StatusUnauthorized)
					return
				}
				user, err := deps.IdentityStore.UpsertUser(r.Context(), hierarchy.UpsertUserInput{Subject: principal.Subject, DisplayName: principal.DisplayName})
				if err != nil {
					http.Error(w, "failed to resolve user", http.StatusInternalServerError)
					return
				}
				owns, err := deps.IdentityStore.UserOwnsPlayer(r.Context(), user.ID, *input.PlayerID)
				if err != nil {
					http.Error(w, "failed to verify player ownership", http.StatusInternalServerError)
					return
				}
				if !owns {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
				input.Subject = principal.Subject
			}
			link, err := deps.PlatformStore.UnlinkPlatformAccount(r.Context(), input)
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

			status, err := deps.EligibilityStore.GetEligibilityStatus(r.Context(), subject)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, status)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func authorizeScopedRoleMutation(w http.ResponseWriter, r *http.Request, deps Dependencies, tokenValidator auth.TokenValidator, evaluator authz.Evaluator, cookieName, secret string, scope hierarchy.HierarchyScope, action authz.Action) (int64, bool) {
	principal, ok := readRequestPrincipal(r, cookieName, secret, tokenValidator, deps.APITokenStore)
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return 0, false
	}
	user, err := deps.IdentityStore.UpsertUser(r.Context(), hierarchy.UpsertUserInput{Subject: principal.Subject, DisplayName: principal.DisplayName})
	if err != nil {
		http.Error(w, "failed to resolve user", http.StatusInternalServerError)
		return 0, false
	}
	actorPlayerID, err := deps.IdentityStore.GetUserPlayerIDForGame(r.Context(), user.ID, scope.GameID)
	if err != nil {
		handleHierarchyError(w, err)
		return 0, false
	}
	if !allowedInScope(w, r, deps, evaluator, principal, user.ID, scope, authz.ResourceRoleAssignment, action) {
		return 0, false
	}
	return actorPlayerID, true
}

func authorizeScopedAction(w http.ResponseWriter, r *http.Request, deps Dependencies, tokenValidator auth.TokenValidator, evaluator authz.Evaluator, cookieName, secret string, scope hierarchy.HierarchyScope, resource authz.Resource, action authz.Action) bool {
	principal, ok := readRequestPrincipal(r, cookieName, secret, tokenValidator, deps.APITokenStore)
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return false
	}
	user, err := deps.IdentityStore.UpsertUser(r.Context(), hierarchy.UpsertUserInput{Subject: principal.Subject, DisplayName: principal.DisplayName})
	if err != nil {
		http.Error(w, "failed to resolve user", http.StatusInternalServerError)
		return false
	}
	return allowedInScope(w, r, deps, evaluator, principal, user.ID, scope, resource, action)
}

func allowedInScope(w http.ResponseWriter, r *http.Request, deps Dependencies, evaluator authz.Evaluator, principal auth.SessionPrincipal, userID int64, scope hierarchy.HierarchyScope, resource authz.Resource, action authz.Action) bool {
	scopedRoles, err := deps.RoleStore.ResolveScopedRoles(r.Context(), hierarchy.ResolveScopedRolesInput{UserID: userID, GameID: scope.GameID, FranchiseID: scope.FranchiseID, ClubID: scope.ClubID, TeamID: scope.TeamID})
	if err != nil {
		http.Error(w, "failed to resolve scoped roles", http.StatusInternalServerError)
		return false
	}
	if !evaluator.AllowedInContext(globalAuthorityRoles(principal.Roles), scopedRoles, resource, action, authz.AuthzContext{FranchiseID: scope.FranchiseID, ClubID: scope.ClubID, TeamID: scope.TeamID}) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func globalAuthorityRoles(roles []string) []string {
	result := make([]string, 0, len(roles))
	for _, role := range roles {
		if role == "admin" || role == "operator" {
			result = append(result, role)
		}
	}
	return result
}
