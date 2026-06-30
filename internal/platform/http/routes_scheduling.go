package http

import (
	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
	"net/http"
)

func (r routeRegistrar) registerSchedulingRoutes(mux *http.ServeMux) {
	deps := r.deps

	mux.HandleFunc("/v1/seasons", func(w http.ResponseWriter, r *http.Request) {
		if deps.SchedulingStore == nil {
			http.Error(w, "scheduling store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			seasons, err := deps.SchedulingStore.ListSeasons(r.Context())
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
			season, err := deps.SchedulingStore.CreateSeason(r.Context(), input)
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
		if deps.SchedulingStore == nil {
			http.Error(w, "scheduling store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			groups, err := deps.SchedulingStore.ListScheduleGroups(r.Context())
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
			group, err := deps.SchedulingStore.CreateScheduleGroup(r.Context(), input)
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
		if deps.SchedulingStore == nil {
			http.Error(w, "scheduling store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			fixtures, err := deps.SchedulingStore.ListFixtures(r.Context())
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
			fixture, err := deps.SchedulingStore.CreateFixture(r.Context(), input)
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
		if deps.SchedulingStore == nil {
			http.Error(w, "scheduling store unavailable", http.StatusServiceUnavailable)
			return
		}

		switch r.Method {
		case http.MethodGet:
			matches, err := deps.SchedulingStore.ListMatches(r.Context())
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
			match, err := deps.SchedulingStore.CreateMatch(r.Context(), input)
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
		if deps.SchedulingStore == nil {
			http.Error(w, "scheduling store unavailable", http.StatusServiceUnavailable)
			return
		}
		id, err := parsePathID(r)
		if err != nil {
			http.Error(w, "invalid fixture id", http.StatusBadRequest)
			return
		}
		fixture, err := deps.SchedulingStore.GetFixture(r.Context(), id)
		if err != nil {
			handleHierarchyError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, fixture)
	})
}
