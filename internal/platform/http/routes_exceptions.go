package http

import (
	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
	"net/http"
)

func (r routeRegistrar) registerExceptionRoutes(mux *http.ServeMux) {
	deps := r.deps

	mux.HandleFunc("/v1/exceptions/report", func(w http.ResponseWriter, r *http.Request) {
		if deps.ExceptionStore == nil {
			http.Error(w, "exception store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input hierarchy.ReportExceptionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			ticket, err := deps.ExceptionStore.ReportException(r.Context(), input)
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
		if deps.ExceptionStore == nil {
			http.Error(w, "exception store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			tickets, err := deps.ExceptionStore.ListOperatorInbox(r.Context())
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
		if deps.ExceptionStore == nil {
			http.Error(w, "exception store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input hierarchy.TriageExceptionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			ticket, err := deps.ExceptionStore.TriageException(r.Context(), input)
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
		if deps.ExceptionStore == nil {
			http.Error(w, "exception store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input hierarchy.ResolveExceptionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			ticket, err := deps.ExceptionStore.ResolveException(r.Context(), input)
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
		if deps.ExceptionStore == nil {
			http.Error(w, "exception store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			actions, err := deps.ExceptionStore.ListExceptionActions(r.Context())
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
		if deps.ExceptionStore == nil {
			http.Error(w, "exception store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodGet:
			metrics, err := deps.ExceptionStore.GetExceptionMetrics(r.Context())
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
		if deps.ExceptionStore == nil {
			http.Error(w, "exception store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input hierarchy.EvaluateSchedulingExceptionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			result, err := deps.ExceptionStore.EvaluateSchedulingException(r.Context(), input)
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
		if deps.ExceptionStore == nil {
			http.Error(w, "exception store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input hierarchy.EvaluateNoShowExceptionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			result, err := deps.ExceptionStore.EvaluateNoShowException(r.Context(), input)
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
		if deps.ExceptionStore == nil {
			http.Error(w, "exception store unavailable", http.StatusServiceUnavailable)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var input hierarchy.EvaluateReplayDisputeExceptionInput
			if err := decodeJSON(r, &input); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			result, err := deps.ExceptionStore.EvaluateReplayDisputeException(r.Context(), input)
			if err != nil {
				handleHierarchyError(w, err)
				return
			}
			writeJSON(w, http.StatusOK, result)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
