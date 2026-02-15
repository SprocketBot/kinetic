package http

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
	"github.com/sprocketbot/sprocket-v3/internal/platform/config"
)

type probeResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

type adminPingResponse struct {
	Status  string `json:"status"`
	Subject string `json:"subject"`
}

type fakeHierarchyStore struct {
	leagueToReturn        hierarchy.League
	franchiseToReturn     hierarchy.Franchise
	clubToReturn          hierarchy.Club
	teamToReturn          hierarchy.Team
	playerToReturn        hierarchy.Player
	membershipToReturn    hierarchy.RosterMembership
	queueToReturn         hierarchy.Queue
	queueEntryToReturn    hierarchy.QueueEntry
	scrimToReturn         hierarchy.Scrim
	submissionToReturn    hierarchy.ResultSubmission
	ingestionToReturn     hierarchy.ReplayIngestionResult
	seasonToReturn        hierarchy.Season
	groupToReturn         hierarchy.ScheduleGroup
	fixtureToReturn       hierarchy.Fixture
	matchToReturn         hierarchy.Match
	leaguesToList         []hierarchy.League
	franchisesToList      []hierarchy.Franchise
	clubsToList           []hierarchy.Club
	teamsToList           []hierarchy.Team
	playersToList         []hierarchy.Player
	membershipsToList     []hierarchy.RosterMembership
	queuesToList          []hierarchy.Queue
	queueEntriesToList    []hierarchy.QueueEntry
	scrimsToList          []hierarchy.Scrim
	runsToList            []hierarchy.PromotionProcessingRun
	ratingsToList         []hierarchy.PlayerRating
	decisionsToList       []hierarchy.MatchmakingDecision
	submissionsToList     []hierarchy.ResultSubmission
	replayEvidenceToList  []hierarchy.ReplayEvidence
	replayParseRunsToList []hierarchy.ReplayParseRun
	replayLinksToList     []hierarchy.ResultSubmissionReplayLink
	seasonsToList         []hierarchy.Season
	groupsToList          []hierarchy.ScheduleGroup
	fixturesToList        []hierarchy.Fixture
	matchesToList         []hierarchy.Match
	createLeagueErr       error
	createFranchiseErr    error
	createClubErr         error
	createTeamErr         error
	createPlayerErr       error
	createMemberErr       error
	createQueueErr        error
	enqueueTeamErr        error
	leaveQueueErr         error
	advanceStageErr       error
	createScrimErr        error
	updateScrimErr        error
	promoteQueueErr       error
	processPromoteErr     error
	createSubmissionErr   error
	ratifySubmissionErr   error
	rejectSubmissionErr   error
	ingestReplayErr       error
	createSeasonErr       error
	createGroupErr        error
	createFixtureErr      error
	createMatchErr        error
}

func (f *fakeHierarchyStore) CreateLeague(_ context.Context, _ hierarchy.CreateLeagueInput) (hierarchy.League, error) {
	return f.leagueToReturn, f.createLeagueErr
}
func (f *fakeHierarchyStore) ListLeagues(_ context.Context) ([]hierarchy.League, error) {
	return f.leaguesToList, nil
}
func (f *fakeHierarchyStore) CreateFranchise(_ context.Context, _ hierarchy.CreateFranchiseInput) (hierarchy.Franchise, error) {
	return f.franchiseToReturn, f.createFranchiseErr
}
func (f *fakeHierarchyStore) ListFranchises(_ context.Context) ([]hierarchy.Franchise, error) {
	return f.franchisesToList, nil
}
func (f *fakeHierarchyStore) CreateClub(_ context.Context, _ hierarchy.CreateClubInput) (hierarchy.Club, error) {
	return f.clubToReturn, f.createClubErr
}
func (f *fakeHierarchyStore) ListClubs(_ context.Context) ([]hierarchy.Club, error) {
	return f.clubsToList, nil
}
func (f *fakeHierarchyStore) CreateTeam(_ context.Context, _ hierarchy.CreateTeamInput) (hierarchy.Team, error) {
	return f.teamToReturn, f.createTeamErr
}
func (f *fakeHierarchyStore) ListTeams(_ context.Context) ([]hierarchy.Team, error) {
	return f.teamsToList, nil
}
func (f *fakeHierarchyStore) CreatePlayer(_ context.Context, _ hierarchy.CreatePlayerInput) (hierarchy.Player, error) {
	return f.playerToReturn, f.createPlayerErr
}
func (f *fakeHierarchyStore) ListPlayers(_ context.Context) ([]hierarchy.Player, error) {
	return f.playersToList, nil
}
func (f *fakeHierarchyStore) CreateRosterMembership(_ context.Context, _ hierarchy.CreateRosterMembershipInput) (hierarchy.RosterMembership, error) {
	return f.membershipToReturn, f.createMemberErr
}
func (f *fakeHierarchyStore) ListRosterMemberships(_ context.Context) ([]hierarchy.RosterMembership, error) {
	return f.membershipsToList, nil
}
func (f *fakeHierarchyStore) CreateQueue(_ context.Context, _ hierarchy.CreateQueueInput) (hierarchy.Queue, error) {
	return f.queueToReturn, f.createQueueErr
}
func (f *fakeHierarchyStore) ListQueues(_ context.Context) ([]hierarchy.Queue, error) {
	return f.queuesToList, nil
}
func (f *fakeHierarchyStore) EnqueueTeam(_ context.Context, _ hierarchy.EnqueueTeamInput) (hierarchy.QueueEntry, error) {
	return f.queueEntryToReturn, f.enqueueTeamErr
}
func (f *fakeHierarchyStore) LeaveQueue(_ context.Context, _ hierarchy.LeaveQueueInput) (hierarchy.QueueEntry, error) {
	return f.queueEntryToReturn, f.leaveQueueErr
}
func (f *fakeHierarchyStore) AdvanceQueueEntryStage(_ context.Context, _ hierarchy.AdvanceQueueEntryStageInput) (hierarchy.QueueEntry, error) {
	return f.queueEntryToReturn, f.advanceStageErr
}
func (f *fakeHierarchyStore) ListActiveQueueEntries(_ context.Context) ([]hierarchy.QueueEntry, error) {
	return f.queueEntriesToList, nil
}
func (f *fakeHierarchyStore) CreateScrim(_ context.Context, _ hierarchy.CreateScrimInput) (hierarchy.Scrim, error) {
	return f.scrimToReturn, f.createScrimErr
}
func (f *fakeHierarchyStore) ListScrims(_ context.Context) ([]hierarchy.Scrim, error) {
	return f.scrimsToList, nil
}
func (f *fakeHierarchyStore) UpdateScrimState(_ context.Context, _ hierarchy.UpdateScrimStateInput) (hierarchy.Scrim, error) {
	return f.scrimToReturn, f.updateScrimErr
}
func (f *fakeHierarchyStore) PromoteQueueToScrim(_ context.Context, _ hierarchy.PromoteQueueToScrimInput) (hierarchy.Scrim, error) {
	return f.scrimToReturn, f.promoteQueueErr
}
func (f *fakeHierarchyStore) ProcessQueuePromotions(_ context.Context, _ hierarchy.ProcessQueuePromotionsInput) (hierarchy.ProcessQueuePromotionsResult, error) {
	return hierarchy.ProcessQueuePromotionsResult{
		ProcessedQueues:   1,
		PromotionsCreated: 1,
		Conflicts:         0,
	}, f.processPromoteErr
}
func (f *fakeHierarchyStore) ListPromotionProcessingRuns(_ context.Context) ([]hierarchy.PromotionProcessingRun, error) {
	return f.runsToList, nil
}
func (f *fakeHierarchyStore) ListPlayerRatings(_ context.Context) ([]hierarchy.PlayerRating, error) {
	return f.ratingsToList, nil
}
func (f *fakeHierarchyStore) ListMatchmakingDecisions(_ context.Context) ([]hierarchy.MatchmakingDecision, error) {
	return f.decisionsToList, nil
}
func (f *fakeHierarchyStore) CreateResultSubmission(_ context.Context, _ hierarchy.CreateResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	return f.submissionToReturn, f.createSubmissionErr
}
func (f *fakeHierarchyStore) ListResultSubmissions(_ context.Context) ([]hierarchy.ResultSubmission, error) {
	return f.submissionsToList, nil
}
func (f *fakeHierarchyStore) RatifyResultSubmission(_ context.Context, _ hierarchy.RatifyResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	return f.submissionToReturn, f.ratifySubmissionErr
}
func (f *fakeHierarchyStore) RejectResultSubmission(_ context.Context, _ hierarchy.RejectResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	return f.submissionToReturn, f.rejectSubmissionErr
}
func (f *fakeHierarchyStore) IngestReplayEvidence(_ context.Context, _ hierarchy.IngestReplayEvidenceInput) (hierarchy.ReplayIngestionResult, error) {
	return f.ingestionToReturn, f.ingestReplayErr
}
func (f *fakeHierarchyStore) ListReplayEvidence(_ context.Context) ([]hierarchy.ReplayEvidence, error) {
	return f.replayEvidenceToList, nil
}
func (f *fakeHierarchyStore) ListReplayParseRuns(_ context.Context) ([]hierarchy.ReplayParseRun, error) {
	return f.replayParseRunsToList, nil
}
func (f *fakeHierarchyStore) ListResultSubmissionReplayLinks(_ context.Context) ([]hierarchy.ResultSubmissionReplayLink, error) {
	return f.replayLinksToList, nil
}
func (f *fakeHierarchyStore) CreateSeason(_ context.Context, _ hierarchy.CreateSeasonInput) (hierarchy.Season, error) {
	return f.seasonToReturn, f.createSeasonErr
}
func (f *fakeHierarchyStore) ListSeasons(_ context.Context) ([]hierarchy.Season, error) {
	return f.seasonsToList, nil
}
func (f *fakeHierarchyStore) CreateScheduleGroup(_ context.Context, _ hierarchy.CreateScheduleGroupInput) (hierarchy.ScheduleGroup, error) {
	return f.groupToReturn, f.createGroupErr
}
func (f *fakeHierarchyStore) ListScheduleGroups(_ context.Context) ([]hierarchy.ScheduleGroup, error) {
	return f.groupsToList, nil
}
func (f *fakeHierarchyStore) CreateFixture(_ context.Context, _ hierarchy.CreateFixtureInput) (hierarchy.Fixture, error) {
	return f.fixtureToReturn, f.createFixtureErr
}
func (f *fakeHierarchyStore) ListFixtures(_ context.Context) ([]hierarchy.Fixture, error) {
	return f.fixturesToList, nil
}
func (f *fakeHierarchyStore) CreateMatch(_ context.Context, _ hierarchy.CreateMatchInput) (hierarchy.Match, error) {
	return f.matchToReturn, f.createMatchErr
}
func (f *fakeHierarchyStore) ListMatches(_ context.Context) ([]hierarchy.Match, error) {
	return f.matchesToList, nil
}

func TestHealthEndpoint(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	assertProbePayload(t, rr.Body, "ok")
}

func TestReadyEndpoint(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	assertProbePayload(t, rr.Body, "ready")
}

func TestAdminPingRequiresAuthentication(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/ping", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}

func TestAdminPingRejectsInsufficientRole(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/ping", nil)
	req.Header.Set("Authorization", "Bearer local:bob:observer")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

func TestAdminPingAllowsAdmin(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/ping", nil)
	req.Header.Set("Authorization", "Bearer local:alice:admin")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var payload adminPingResponse
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode admin ping payload: %v", err)
	}

	if payload.Subject != "alice" {
		t.Fatalf("expected subject alice, got %s", payload.Subject)
	}
}

func TestLeaguesEndpointUnavailableWithoutStore(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/leagues", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rr.Code)
	}
}

func TestCreateLeagueSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		leagueToReturn: hierarchy.League{
			ID:        1,
			Name:      "MLE",
			Slug:      "mle",
			IsActive:  true,
			CreatedAt: now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/leagues", strings.NewReader(`{"name":"MLE","slug":"mle"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestCreateLeagueValidationError(t *testing.T) {
	store := &fakeHierarchyStore{
		createLeagueErr: hierarchy.ErrInvalidInput,
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/leagues", strings.NewReader(`{"name":"MLE","slug":"BAD-SLUG"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCreateTeamSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		teamToReturn: hierarchy.Team{
			ID:        1,
			ClubID:    1,
			Name:      "Team Alpha",
			Slug:      "team-alpha",
			IsActive:  true,
			CreatedAt: now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/teams", strings.NewReader(`{"clubId":1,"name":"Team Alpha","slug":"team-alpha"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestCreatePlayerSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		playerToReturn: hierarchy.Player{
			ID:          1,
			DisplayName: "Player One",
			Slug:        "player-one",
			IsActive:    true,
			CreatedAt:   now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/players", strings.NewReader(`{"displayName":"Player One","slug":"player-one"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestCreateRosterMembershipSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		membershipToReturn: hierarchy.RosterMembership{
			ID:        1,
			PlayerID:  10,
			TeamID:    20,
			IsActive:  true,
			CreatedAt: now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/roster-memberships", strings.NewReader(`{"playerId":10,"teamId":20}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestCreateQueueSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		queueToReturn: hierarchy.Queue{
			ID:        1,
			Name:      "3v3 Ranked",
			Slug:      "3v3-ranked",
			IsActive:  true,
			CreatedAt: now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/queues", strings.NewReader(`{"name":"3v3 Ranked","slug":"3v3-ranked"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestJoinQueueSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		queueEntryToReturn: hierarchy.QueueEntry{
			ID:        1,
			QueueID:   10,
			TeamID:    20,
			IsActive:  true,
			CreatedAt: now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/queue-entries", strings.NewReader(`{"queueId":10,"teamId":20}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestCreateSeasonSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		seasonToReturn: hierarchy.Season{
			ID:        1,
			Name:      "Season 1",
			Slug:      "season-1",
			IsActive:  true,
			CreatedAt: now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/seasons", strings.NewReader(`{"name":"Season 1","slug":"season-1"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestCreateMatchReadySuccess(t *testing.T) {
	scheduled := time.Now().UTC().Add(24 * time.Hour)
	homeRatified := scheduled.Add(-2 * time.Hour)
	awayRatified := scheduled.Add(-1 * time.Hour)
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		matchToReturn: hierarchy.Match{
			ID:                 1,
			FixtureID:          10,
			HomeTeamID:         20,
			AwayTeamID:         30,
			State:              "ready",
			ScheduledFor:       &scheduled,
			HomeTimeRatifiedAt: &homeRatified,
			AwayTimeRatifiedAt: &awayRatified,
			CreatedAt:          now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/matches", strings.NewReader(
		`{"fixtureId":10,"homeTeamId":20,"awayTeamId":30,"state":"ready","scheduledFor":"2030-01-01T10:00:00Z","homeTimeRatifiedAt":"2030-01-01T08:00:00Z","awayTimeRatifiedAt":"2030-01-01T09:00:00Z"}`,
	))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestPromoteQueueToScrimSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		scrimToReturn: hierarchy.Scrim{
			ID:         1,
			QueueID:    10,
			HomeTeamID: 20,
			AwayTeamID: 30,
			State:      "created",
			CreatedAt:  now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/scrim-promotions", strings.NewReader(`{"queueId":10}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestUpdateScrimStateSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		scrimToReturn: hierarchy.Scrim{
			ID:         1,
			QueueID:    10,
			HomeTeamID: 20,
			AwayTeamID: 30,
			State:      "in_progress",
			CreatedAt:  now.Add(-10 * time.Minute),
			StartedAt:  &now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPatch, "/v1/scrims", strings.NewReader(`{"scrimId":1,"state":"in_progress"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestProcessQueuePromotionsSuccess(t *testing.T) {
	store := &fakeHierarchyStore{}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/scrim-promotions/process", strings.NewReader(`{"queueId":0}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestListPromotionProcessingRunsSuccess(t *testing.T) {
	now := time.Now().UTC()
	queueID := int64(10)
	store := &fakeHierarchyStore{
		runsToList: []hierarchy.PromotionProcessingRun{
			{
				ID:                1,
				QueueID:           &queueID,
				ProcessedQueues:   1,
				PromotionsCreated: 1,
				Conflicts:         0,
				DurationMs:        12,
				CreatedAt:         now,
			},
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/promotion-processing-runs", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestCreateResultSubmissionSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		submissionToReturn: hierarchy.ResultSubmission{
			ID:                1,
			ContextType:       "scrim",
			ContextID:         10,
			SubmittedByTeamID: 20,
			HomeTeamID:        20,
			AwayTeamID:        30,
			WinningTeamID:     20,
			LosingTeamID:      30,
			State:             "pending",
			PayloadJSON:       []byte(`{"score":"3-1"}`),
			CreatedAt:         now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/result-submissions", strings.NewReader(`{"contextType":"scrim","contextId":10,"submittedByTeamId":20,"winningTeamId":20,"losingTeamId":30,"payloadJson":{"score":"3-1"}}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestRatifyResultSubmissionSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		submissionToReturn: hierarchy.ResultSubmission{
			ID:             1,
			ContextType:    "scrim",
			ContextID:      10,
			State:          "ratified",
			HomeRatifiedAt: &now,
			AwayRatifiedAt: &now,
			CreatedAt:      now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/result-submission-ratifications", strings.NewReader(`{"submissionId":1,"teamId":20}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestIngestReplayEvidenceSuccess(t *testing.T) {
	now := time.Now().UTC()
	submissionID := int64(1)
	store := &fakeHierarchyStore{
		ingestionToReturn: hierarchy.ReplayIngestionResult{
			Evidence: hierarchy.ReplayEvidence{
				ID:                1,
				ContextType:       "scrim",
				ContextID:         10,
				SubmittedByTeamID: 20,
				ReplaySHA256:      "abc123",
				ContentSizeBytes:  42,
				StorageRef:        "inline-sha256:abc123",
				State:             "parsed",
				CreatedAt:         now,
			},
			ParseRun: hierarchy.ReplayParseRun{
				ID:                 1,
				ReplayEvidenceID:   1,
				ParserName:         "sprocket-rl-parser",
				ParserVersion:      "v0.1.0",
				ParserConfigDigest: "cfg-001",
				Status:             "parsed",
				OutputJSON:         []byte(`{"goals":4}`),
				CreatedAt:          now,
			},
			Duplicate:          false,
			LinkedSubmissionID: &submissionID,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/replay-evidence", strings.NewReader(`{"contextType":"scrim","contextId":10,"submittedByTeamId":20,"replayBody":"fake-bytes","parserName":"sprocket-rl-parser","parserVersion":"v0.1.0","parserConfigDigest":"cfg-001","resultSubmissionId":1,"parseOutputJson":{"goals":4}}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestListReplayParseRunsSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		replayParseRunsToList: []hierarchy.ReplayParseRun{
			{
				ID:                 1,
				ReplayEvidenceID:   1,
				ParserName:         "sprocket-rl-parser",
				ParserVersion:      "v0.1.0",
				ParserConfigDigest: "cfg-001",
				Status:             "parsed",
				OutputJSON:         []byte(`{"goals":4}`),
				CreatedAt:          now,
			},
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/replay-parse-runs", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func assertProbePayload(t *testing.T, body io.Reader, expectedStatus string) {
	t.Helper()

	var payload probeResponse
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode JSON payload: %v", err)
	}

	if payload.Status != expectedStatus {
		t.Fatalf("expected status %q, got %q", expectedStatus, payload.Status)
	}

	if payload.Service != "sprocket-v3-api" {
		t.Fatalf("unexpected service value %q", payload.Service)
	}
}
