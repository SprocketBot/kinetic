package http

import (
	"fmt"
	"github.com/kineticbot/kinetic-v3/internal/domain/authz"
	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
	"github.com/kineticbot/kinetic-v3/internal/platform/auth"
	"net/http"
)

func (r routeRegistrar) registerResultsAndReplayRoutes(mux *http.ServeMux) {
	deps := r.deps
	tokenValidator := r.tokenValidator
	evaluator := r.evaluator
	sessionCookieName := r.sessionCookieName
	sessionSecret := r.sessionSecret

	mux.HandleFunc("/v1/result-submissions", func(w http.ResponseWriter, r *http.Request) {
		if deps.HierarchyStore == nil {
			http.Error(w, "hierarchy store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			submissions, err := deps.ResultStore.ListResultSubmissions(r.Context())
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
			if principal, ok := readRequestPrincipal(r, sessionCookieName, sessionSecret, tokenValidator, deps.APITokenStore); ok {
				input.SubmittedBySubject = principal.Subject
				input.SubmittedByDisplayName = principal.DisplayName
			} else {
				input.SubmittedBySubject = fmt.Sprintf("team:%d:submission", input.SubmittedByTeamID)
				input.SubmittedByDisplayName = fmt.Sprintf("Team %d submitter", input.SubmittedByTeamID)
			}
			submission, err := deps.ResultStore.CreateResultSubmission(r.Context(), input)
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
			overrides, err := deps.ResultStore.ListResultOverrides(r.Context())
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
			existing, err := deps.ResultStore.GetResultSubmission(r.Context(), input.SubmissionID)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			scope, err := deps.RoleStore.ResolveTeamScope(r.Context(), existing.HomeTeamID)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			principal, ok := authorizeScopedPrincipalAction(w, r, deps, tokenValidator, evaluator, sessionCookieName, sessionSecret, scope, authz.ResourceResultOverride, authz.ActionCreate)
			if !ok {
				return
			}
			input.Actor = principal.Subject
			submission, err := deps.ResultStore.OverrideResultSubmission(r.Context(), input)
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
			scope, err := deps.RoleStore.ResolveTeamScope(r.Context(), input.TeamID)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			principal, ok := authorizeScopedPrincipalAction(w, r, deps, tokenValidator, evaluator, sessionCookieName, sessionSecret, scope, authz.ResourceResultSubmission, authz.ActionRatify)
			if !ok {
				return
			}
			input.RatifiedBySubject = principal.Subject
			input.RatifiedByDisplayName = principal.DisplayName
			submission, err := deps.ResultStore.RatifyResultSubmission(r.Context(), input)
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
			scope, err := deps.RoleStore.ResolveTeamScope(r.Context(), input.TeamID)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			if _, ok := authorizeScopedPrincipalAction(w, r, deps, tokenValidator, evaluator, sessionCookieName, sessionSecret, scope, authz.ResourceResultSubmission, authz.ActionReject); !ok {
				return
			}
			submission, err := deps.ResultStore.RejectResultSubmission(r.Context(), input)
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
			evidence, err := deps.ReplayStore.ListReplayEvidence(r.Context())
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
			result, err := deps.ReplayStore.IngestReplayEvidence(r.Context(), input)
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
			runs, err := deps.ReplayStore.ListReplayParseRuns(r.Context())
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
			links, err := deps.ReplayStore.ListResultSubmissionReplayLinks(r.Context())
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
}

func authorizeScopedPrincipalAction(w http.ResponseWriter, r *http.Request, deps Dependencies, tokenValidator auth.TokenValidator, evaluator authz.Evaluator, cookieName, secret string, scope hierarchy.HierarchyScope, resource authz.Resource, action authz.Action) (auth.SessionPrincipal, bool) {
	principal, ok := readRequestPrincipal(r, cookieName, secret, tokenValidator, deps.APITokenStore)
	if !ok {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return auth.SessionPrincipal{}, false
	}
	user, err := deps.IdentityStore.UpsertUser(r.Context(), hierarchy.UpsertUserInput{Subject: principal.Subject, DisplayName: principal.DisplayName})
	if err != nil {
		http.Error(w, "failed to resolve user", http.StatusInternalServerError)
		return auth.SessionPrincipal{}, false
	}
	if !allowedInScope(w, r, deps, evaluator, principal, user.ID, scope, resource, action) {
		return auth.SessionPrincipal{}, false
	}
	return principal, true
}
