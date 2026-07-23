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

	"github.com/kineticbot/kinetic-v3/internal/domain/apitoken"
	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
	"github.com/kineticbot/kinetic-v3/internal/domain/replaystats"
	"github.com/kineticbot/kinetic-v3/internal/platform/config"
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
	leagueToReturn           hierarchy.League
	franchiseToReturn        hierarchy.Franchise
	clubToReturn             hierarchy.Club
	teamToReturn             hierarchy.Team
	playerToReturn           hierarchy.Player
	userToReturn             hierarchy.User
	gameToReturn             hierarchy.Game
	gamesToList              []hierarchy.Game
	userPlayerToReturn       hierarchy.UserPlayer
	userPlayersToList        []hierarchy.UserPlayer
	createUserPlayerInput    hierarchy.CreateUserPlayerInput
	linkPlatformInput        hierarchy.LinkPlatformAccountInput
	ownsPlayer               bool
	membershipToReturn       hierarchy.RosterMembership
	roleAssignmentToReturn   hierarchy.RoleAssignment
	queueToReturn            hierarchy.Queue
	eligibilityToReturn      hierarchy.EligibilityStatus
	queueEntryToReturn       hierarchy.QueueEntry
	scrimToReturn            hierarchy.Scrim
	submissionToReturn       hierarchy.ResultSubmission
	ingestionToReturn        hierarchy.ReplayIngestionResult
	seasonToReturn           hierarchy.Season
	groupToReturn            hierarchy.ScheduleGroup
	fixtureToReturn          hierarchy.Fixture
	matchToReturn            hierarchy.Match
	leaguesToList            []hierarchy.League
	franchisesToList         []hierarchy.Franchise
	clubsToList              []hierarchy.Club
	teamsToList              []hierarchy.Team
	playersToList            []hierarchy.Player
	membershipsToList        []hierarchy.RosterMembership
	roleAssignmentsToList    []hierarchy.RoleAssignment
	queuesToList             []hierarchy.Queue
	platformLinksToList      []hierarchy.PlatformAccountLink
	queueEntriesToList       []hierarchy.QueueEntry
	queueBansToList          []hierarchy.QueueBan
	scrimsToList             []hierarchy.Scrim
	runsToList               []hierarchy.PromotionProcessingRun
	ratingsToList            []hierarchy.PlayerRating
	ratingAdjustmentsToList  []hierarchy.RatingAdjustment
	decisionsToList          []hierarchy.MatchmakingDecision
	submissionsToList        []hierarchy.ResultSubmission
	resultOverridesToList    []hierarchy.ResultOverride
	replayEvidenceToList     []hierarchy.ReplayEvidence
	replayParseRunsToList    []hierarchy.ReplayParseRun
	replayLinksToList        []hierarchy.ResultSubmissionReplayLink
	exceptionToReturn        hierarchy.ExceptionTicket
	exceptionResultToReturn  hierarchy.ExceptionAutomationResult
	exceptionActionsToList   []hierarchy.ExceptionAction
	exceptionMetricsToReturn hierarchy.ExceptionMetrics
	exceptionsToList         []hierarchy.ExceptionTicket
	seasonsToList            []hierarchy.Season
	groupsToList             []hierarchy.ScheduleGroup
	fixturesToList           []hierarchy.Fixture
	matchesToList            []hierarchy.Match
	createLeagueErr          error
	createFranchiseErr       error
	createClubErr            error
	createTeamErr            error
	createPlayerErr          error
	createMemberErr          error
	assignRoleErr            error
	revokeRoleErr            error
	createQueueErr           error
	eligibilityErr           error
	linkPlatformErr          error
	unlinkPlatformErr        error
	enqueueTeamErr           error
	leaveQueueErr            error
	banQueuePlayerErr        error
	unbanQueuePlayerErr      error
	advanceStageErr          error
	createScrimErr           error
	updateScrimErr           error
	promoteQueueErr          error
	processPromoteErr        error
	createSubmissionErr      error
	overrideSubmissionErr    error
	ratifySubmissionErr      error
	rejectSubmissionErr      error
	adjustRatingErr          error
	ingestReplayErr          error
	reportExceptionErr       error
	triageExceptionErr       error
	resolveExceptionErr      error
	schedulingEvalErr        error
	noShowEvalErr            error
	replayDisputeEvalErr     error
	createSeasonErr          error
	createGroupErr           error
	createFixtureErr         error
	createMatchErr           error
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
func (f *fakeHierarchyStore) UpsertUser(_ context.Context, _ hierarchy.UpsertUserInput) (hierarchy.User, error) {
	return f.userToReturn, nil
}
func (f *fakeHierarchyStore) CreateGame(_ context.Context, _ hierarchy.CreateGameInput) (hierarchy.Game, error) {
	return f.gameToReturn, nil
}
func (f *fakeHierarchyStore) ListGames(_ context.Context) ([]hierarchy.Game, error) {
	return f.gamesToList, nil
}
func (f *fakeHierarchyStore) CreateUserPlayer(_ context.Context, input hierarchy.CreateUserPlayerInput) (hierarchy.UserPlayer, error) {
	f.createUserPlayerInput = input
	return f.userPlayerToReturn, nil
}
func (f *fakeHierarchyStore) ListUserPlayers(_ context.Context, _ int64) ([]hierarchy.UserPlayer, error) {
	return f.userPlayersToList, nil
}
func (f *fakeHierarchyStore) UserOwnsPlayer(_ context.Context, _, _ int64) (bool, error) {
	return f.ownsPlayer, nil
}
func (f *fakeHierarchyStore) CreateRosterMembership(_ context.Context, _ hierarchy.CreateRosterMembershipInput) (hierarchy.RosterMembership, error) {
	return f.membershipToReturn, f.createMemberErr
}
func (f *fakeHierarchyStore) ListRosterMemberships(_ context.Context) ([]hierarchy.RosterMembership, error) {
	return f.membershipsToList, nil
}
func (f *fakeHierarchyStore) AssignRole(_ context.Context, _ hierarchy.AssignRoleInput) (hierarchy.RoleAssignment, error) {
	if f.roleAssignmentToReturn.ID != 0 {
		return f.roleAssignmentToReturn, f.assignRoleErr
	}
	if len(f.roleAssignmentsToList) > 0 {
		return f.roleAssignmentsToList[0], f.assignRoleErr
	}
	return hierarchy.RoleAssignment{}, f.assignRoleErr
}
func (f *fakeHierarchyStore) RevokeRole(_ context.Context, _ hierarchy.RevokeRoleInput) (hierarchy.RoleAssignment, error) {
	if f.roleAssignmentToReturn.ID != 0 {
		return f.roleAssignmentToReturn, f.revokeRoleErr
	}
	if len(f.roleAssignmentsToList) > 0 {
		return f.roleAssignmentsToList[0], f.revokeRoleErr
	}
	return hierarchy.RoleAssignment{}, f.revokeRoleErr
}
func (f *fakeHierarchyStore) ListRoleAssignments(_ context.Context) ([]hierarchy.RoleAssignment, error) {
	return f.roleAssignmentsToList, nil
}
func (f *fakeHierarchyStore) ResolveScopedRoles(_ context.Context, _ hierarchy.ResolveScopedRolesInput) ([]string, error) {
	return nil, nil
}
func (f *fakeHierarchyStore) CreateQueue(_ context.Context, _ hierarchy.CreateQueueInput) (hierarchy.Queue, error) {
	return f.queueToReturn, f.createQueueErr
}
func (f *fakeHierarchyStore) ListQueues(_ context.Context) ([]hierarchy.Queue, error) {
	return f.queuesToList, nil
}
func (f *fakeHierarchyStore) LinkPlatformAccount(_ context.Context, input hierarchy.LinkPlatformAccountInput) (hierarchy.PlatformAccountLink, error) {
	f.linkPlatformInput = input
	if len(f.platformLinksToList) > 0 {
		return f.platformLinksToList[0], f.linkPlatformErr
	}
	return hierarchy.PlatformAccountLink{}, f.linkPlatformErr
}
func (f *fakeHierarchyStore) UnlinkPlatformAccount(_ context.Context, _ hierarchy.UnlinkPlatformAccountInput) (hierarchy.PlatformAccountLink, error) {
	if len(f.platformLinksToList) > 0 {
		return f.platformLinksToList[0], f.unlinkPlatformErr
	}
	return hierarchy.PlatformAccountLink{}, f.unlinkPlatformErr
}
func (f *fakeHierarchyStore) ListPlatformAccountLinks(_ context.Context, _ string) ([]hierarchy.PlatformAccountLink, error) {
	return f.platformLinksToList, nil
}
func (f *fakeHierarchyStore) ListPlatformAccountLinksByPlayerID(_ context.Context, _ int64) ([]hierarchy.PlatformAccountLink, error) {
	return f.platformLinksToList, nil
}
func (f *fakeHierarchyStore) GetEligibilityStatus(_ context.Context, _ string) (hierarchy.EligibilityStatus, error) {
	return f.eligibilityToReturn, f.eligibilityErr
}
func (f *fakeHierarchyStore) EnqueueTeam(_ context.Context, _ hierarchy.EnqueueTeamInput) (hierarchy.QueueEntry, error) {
	return f.queueEntryToReturn, f.enqueueTeamErr
}
func (f *fakeHierarchyStore) LeaveQueue(_ context.Context, _ hierarchy.LeaveQueueInput) (hierarchy.QueueEntry, error) {
	return f.queueEntryToReturn, f.leaveQueueErr
}
func (f *fakeHierarchyStore) BanPlayerFromQueue(_ context.Context, _ hierarchy.BanPlayerFromQueueInput) (hierarchy.QueueBan, error) {
	if len(f.queueBansToList) > 0 {
		return f.queueBansToList[0], f.banQueuePlayerErr
	}
	return hierarchy.QueueBan{}, f.banQueuePlayerErr
}
func (f *fakeHierarchyStore) UnbanPlayerFromQueue(_ context.Context, _ hierarchy.UnbanPlayerFromQueueInput) (hierarchy.QueueBan, error) {
	if len(f.queueBansToList) > 0 {
		return f.queueBansToList[0], f.unbanQueuePlayerErr
	}
	return hierarchy.QueueBan{}, f.unbanQueuePlayerErr
}
func (f *fakeHierarchyStore) ListQueueBans(_ context.Context) ([]hierarchy.QueueBan, error) {
	return f.queueBansToList, nil
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
func (f *fakeHierarchyStore) AdjustPlayerRating(_ context.Context, _ hierarchy.AdjustPlayerRatingInput) (hierarchy.PlayerRating, error) {
	if len(f.ratingsToList) > 0 {
		return f.ratingsToList[len(f.ratingsToList)-1], f.adjustRatingErr
	}
	return hierarchy.PlayerRating{}, f.adjustRatingErr
}
func (f *fakeHierarchyStore) ListRatingAdjustments(_ context.Context) ([]hierarchy.RatingAdjustment, error) {
	return f.ratingAdjustmentsToList, nil
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
func (f *fakeHierarchyStore) OverrideResultSubmission(_ context.Context, _ hierarchy.OverrideResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	return f.submissionToReturn, f.overrideSubmissionErr
}
func (f *fakeHierarchyStore) ListResultOverrides(_ context.Context) ([]hierarchy.ResultOverride, error) {
	return f.resultOverridesToList, nil
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
func (f *fakeHierarchyStore) ReportException(_ context.Context, _ hierarchy.ReportExceptionInput) (hierarchy.ExceptionTicket, error) {
	return f.exceptionToReturn, f.reportExceptionErr
}
func (f *fakeHierarchyStore) ListOperatorInbox(_ context.Context) ([]hierarchy.ExceptionTicket, error) {
	return f.exceptionsToList, nil
}
func (f *fakeHierarchyStore) TriageException(_ context.Context, _ hierarchy.TriageExceptionInput) (hierarchy.ExceptionTicket, error) {
	return f.exceptionToReturn, f.triageExceptionErr
}
func (f *fakeHierarchyStore) ResolveException(_ context.Context, _ hierarchy.ResolveExceptionInput) (hierarchy.ExceptionTicket, error) {
	return f.exceptionToReturn, f.resolveExceptionErr
}
func (f *fakeHierarchyStore) ListExceptionActions(_ context.Context) ([]hierarchy.ExceptionAction, error) {
	return f.exceptionActionsToList, nil
}
func (f *fakeHierarchyStore) GetExceptionMetrics(_ context.Context) (hierarchy.ExceptionMetrics, error) {
	return f.exceptionMetricsToReturn, nil
}
func (f *fakeHierarchyStore) EvaluateSchedulingException(_ context.Context, _ hierarchy.EvaluateSchedulingExceptionInput) (hierarchy.ExceptionAutomationResult, error) {
	return f.exceptionResultToReturn, f.schedulingEvalErr
}
func (f *fakeHierarchyStore) EvaluateNoShowException(_ context.Context, _ hierarchy.EvaluateNoShowExceptionInput) (hierarchy.ExceptionAutomationResult, error) {
	return f.exceptionResultToReturn, f.noShowEvalErr
}
func (f *fakeHierarchyStore) EvaluateReplayDisputeException(_ context.Context, _ hierarchy.EvaluateReplayDisputeExceptionInput) (hierarchy.ExceptionAutomationResult, error) {
	return f.exceptionResultToReturn, f.replayDisputeEvalErr
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

func (f *fakeHierarchyStore) GetFixture(_ context.Context, _ int64) (hierarchy.Fixture, error) {
	return hierarchy.Fixture{}, nil
}
func (f *fakeHierarchyStore) GetScrim(_ context.Context, _ int64) (hierarchy.Scrim, error) {
	return hierarchy.Scrim{}, nil
}
func (f *fakeHierarchyStore) GetResultSubmission(_ context.Context, _ int64) (hierarchy.ResultSubmission, error) {
	return hierarchy.ResultSubmission{}, nil
}
func (f *fakeHierarchyStore) ListResultSubmissionsFiltered(_ context.Context, _ hierarchy.ListResultSubmissionsInput) ([]hierarchy.ResultSubmission, error) {
	return nil, nil
}
func (f *fakeHierarchyStore) ResetResultSubmission(_ context.Context, _ hierarchy.ResetResultSubmissionInput) (hierarchy.ResultSubmission, error) {
	return hierarchy.ResultSubmission{}, nil
}
func (f *fakeHierarchyStore) SetPlayerActive(_ context.Context, _ hierarchy.SetPlayerActiveInput) (hierarchy.Player, error) {
	return hierarchy.Player{}, nil
}
func (f *fakeHierarchyStore) CheckInScrim(_ context.Context, _ hierarchy.CheckInScrimInput) (hierarchy.Scrim, error) {
	return hierarchy.Scrim{}, nil
}
func (f *fakeHierarchyStore) ExecutePopTimeout(_ context.Context, _ hierarchy.ExecutePopTimeoutInput) error {
	return nil
}
func (f *fakeHierarchyStore) GetScrimMetrics(_ context.Context) (hierarchy.ScrimMetrics, error) {
	return hierarchy.ScrimMetrics{}, nil
}
func (f *fakeHierarchyStore) GetActiveRosterMembershipByPlayerID(_ context.Context, _ int64) (*hierarchy.RosterMembership, error) {
	return nil, nil
}
func (f *fakeHierarchyStore) ListActiveQueueBansByPlayerID(_ context.Context, _ int64) ([]hierarchy.QueueBan, error) {
	return nil, nil
}
func (f *fakeHierarchyStore) GetActiveQueueEntryByPlayerID(_ context.Context, _ int64) (*hierarchy.QueueEntry, error) {
	return nil, nil
}
func (f *fakeHierarchyStore) GetActiveScrimByPlayerID(_ context.Context, _ int64) (*hierarchy.Scrim, error) {
	return nil, nil
}
func (f *fakeHierarchyStore) TriggerReplayParse(_ context.Context, _, _ int64, _ string) error {
	return nil
}

// fakeReplayStatsStore implements replaystats.Store for tests.
type fakeReplayStatsStore struct {
	statsToReturn       []replaystats.PlayerStatLine
	careerStatsToReturn replaystats.PlayerCareerStats
	statsErr            error
	careerStatsErr      error
}

func (f *fakeReplayStatsStore) ListStatsBySubmission(_ context.Context, _ int64) ([]replaystats.PlayerStatLine, error) {
	return f.statsToReturn, f.statsErr
}
func (f *fakeReplayStatsStore) GetPlayerCareerStats(_ context.Context, _ int64) (replaystats.PlayerCareerStats, error) {
	return f.careerStatsToReturn, f.careerStatsErr
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

func TestSessionLifecycleViaLoginAndCallback(t *testing.T) {
	cfg := config.Config{
		Port:              "8080",
		LogLevel:          "info",
		AuthSessionSecret: "test-secret",
		AuthSessionCookie: "kinetic_session",
		AuthSessionTTL:    "1h",
		WebBaseURL:        "http://localhost:5173",
	}
	srv := New(cfg, slog.Default(), Dependencies{})

	loginReq := httptest.NewRequest(http.MethodGet, "/v1/auth/login?subject=alice&displayName=Alice&roles=league_admin&redirect=http://localhost:5173/app/admin", nil)
	loginRR := httptest.NewRecorder()
	srv.Handler().ServeHTTP(loginRR, loginReq)
	if loginRR.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, loginRR.Code)
	}
	callbackPath := loginRR.Result().Header.Get("Location")
	if !strings.HasPrefix(callbackPath, "/v1/auth/callback?") {
		t.Fatalf("expected callback redirect path, got %s", callbackPath)
	}

	callbackReq := httptest.NewRequest(http.MethodGet, callbackPath, nil)
	callbackRR := httptest.NewRecorder()
	srv.Handler().ServeHTTP(callbackRR, callbackReq)
	if callbackRR.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, callbackRR.Code)
	}
	if callbackRR.Result().Header.Get("Location") != "http://localhost:5173/app/admin" {
		t.Fatalf("unexpected callback redirect target %s", callbackRR.Result().Header.Get("Location"))
	}

	var sessionCookie *http.Cookie
	for _, cookie := range callbackRR.Result().Cookies() {
		if cookie.Name == "kinetic_session" {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie to be issued")
	}

	sessionReq := httptest.NewRequest(http.MethodGet, "/v1/session", nil)
	sessionReq.AddCookie(sessionCookie)
	sessionRR := httptest.NewRecorder()
	srv.Handler().ServeHTTP(sessionRR, sessionReq)
	if sessionRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, sessionRR.Code, sessionRR.Body.String())
	}

	var sessionPayload map[string]any
	if err := json.NewDecoder(sessionRR.Body).Decode(&sessionPayload); err != nil {
		t.Fatalf("failed to decode session payload: %v", err)
	}
	if sessionPayload["subject"] != "alice" {
		t.Fatalf("expected subject alice, got %#v", sessionPayload["subject"])
	}

	logoutReq := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutRR := httptest.NewRecorder()
	srv.Handler().ServeHTTP(logoutRR, logoutReq)
	if logoutRR.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, logoutRR.Code)
	}
}

func TestSessionEndpointRejectsInvalidCookie(t *testing.T) {
	cfg := config.Config{
		Port:              "8080",
		LogLevel:          "info",
		AuthSessionSecret: "test-secret",
		AuthSessionCookie: "kinetic_session",
		AuthSessionTTL:    "1h",
		WebBaseURL:        "http://localhost:5173",
	}
	srv := New(cfg, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/session", nil)
	req.AddCookie(&http.Cookie{Name: "kinetic_session", Value: "bad-token"})
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
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

func TestAssignRoleSuccess(t *testing.T) {
	now := time.Now().UTC()
	clubID := int64(8)
	store := &fakeHierarchyStore{
		roleAssignmentToReturn: hierarchy.RoleAssignment{
			ID:                      1,
			PlayerID:                22,
			Role:                    "gm",
			ClubID:                  &clubID,
			AssignedByActorPlayerID: 11,
			AssignReason:            "season staffing",
			IsActive:                true,
			AssignedAt:              now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/role-assignments", strings.NewReader(
		`{"actorPlayerId":11,"targetPlayerId":22,"role":"gm","clubId":8,"reason":"season staffing"}`,
	))
	req.Header.Set("Authorization", "Bearer local:alice:admin")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestListRoleAssignmentsSuccess(t *testing.T) {
	now := time.Now().UTC()
	teamID := int64(42)
	store := &fakeHierarchyStore{
		roleAssignmentsToList: []hierarchy.RoleAssignment{
			{
				ID:                      1,
				PlayerID:                7,
				Role:                    "captain",
				TeamID:                  &teamID,
				AssignedByActorPlayerID: 5,
				AssignReason:            "team restructure",
				IsActive:                true,
				AssignedAt:              now,
			},
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/role-assignments", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestRevokeRoleSuccess(t *testing.T) {
	now := time.Now().UTC()
	clubID := int64(8)
	store := &fakeHierarchyStore{
		roleAssignmentToReturn: hierarchy.RoleAssignment{
			ID:                      1,
			PlayerID:                22,
			Role:                    "gm",
			ClubID:                  &clubID,
			AssignedByActorPlayerID: 11,
			AssignReason:            "season staffing",
			IsActive:                false,
			AssignedAt:              now.Add(-2 * time.Hour),
			RevokedAt:               &now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/role-assignments/revoke", strings.NewReader(
		`{"actorPlayerId":11,"assignmentId":1,"reason":"role realignment"}`,
	))
	req.Header.Set("Authorization", "Bearer local:alice:admin")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
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

func TestLinkPlatformAccountSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		platformLinksToList: []hierarchy.PlatformAccountLink{
			{
				ID:                1,
				Subject:           "player-1",
				Provider:          "steam",
				ProviderAccountID: "steam-123",
				IsActive:          true,
				LinkedAt:          now,
			},
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/platform-accounts/link", strings.NewReader(`{"subject":"player-1","provider":"steam","providerAccountId":"steam-123","providerAccountName":"Player One"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestLinkPlatformAccountBindsOnlyOwnedPlayer(t *testing.T) {
	playerID := int64(42)
	store := &fakeHierarchyStore{
		userToReturn: hierarchy.User{ID: 7, Subject: "jake", DisplayName: "Jake"},
		ownsPlayer:   true,
		platformLinksToList: []hierarchy.PlatformAccountLink{{
			ID: 1, Subject: "jake", PlayerID: &playerID, Provider: "steam", ProviderAccountID: "steam-42", IsActive: true, LinkedAt: time.Now().UTC(),
		}},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{HierarchyStore: store})
	req := httptest.NewRequest(http.MethodPost, "/v1/platform-accounts/link", strings.NewReader(`{"subject":"other-user","playerId":42,"provider":"steam","providerAccountId":"steam-42","providerAccountName":"Jake"}`))
	req.Header.Set("Authorization", "Bearer local:jake:player")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
	if store.linkPlatformInput.Subject != "jake" || store.linkPlatformInput.PlayerID == nil || *store.linkPlatformInput.PlayerID != playerID {
		t.Fatalf("expected session-owned player link, got %#v", store.linkPlatformInput)
	}
}

func TestLinkPlatformAccountRejectsUnownedPlayer(t *testing.T) {
	playerID := int64(42)
	store := &fakeHierarchyStore{userToReturn: hierarchy.User{ID: 7, Subject: "jake", DisplayName: "Jake"}}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{HierarchyStore: store})
	req := httptest.NewRequest(http.MethodPost, "/v1/platform-accounts/link", strings.NewReader(`{"playerId":42,"provider":"steam","providerAccountId":"steam-42","providerAccountName":"Jake"}`))
	req.Header.Set("Authorization", "Bearer local:jake:player")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d for player %d, got %d body=%s", http.StatusForbidden, playerID, rr.Code, rr.Body.String())
	}
}

func TestListPlatformAccountsSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		platformLinksToList: []hierarchy.PlatformAccountLink{
			{
				ID:                1,
				Subject:           "player-1",
				Provider:          "steam",
				ProviderAccountID: "steam-123",
				IsActive:          true,
				LinkedAt:          now,
			},
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/platform-accounts?subject=player-1", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestGetEligibilityStatusSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		eligibilityToReturn: hierarchy.EligibilityStatus{
			Subject:         "player-1",
			Points:          92,
			ThresholdPoints: 40,
			DecayPerWeek:    10,
			EligibleUntil:   now.AddDate(0, 0, 35),
			EvaluatedAt:     now,
			Projection: []hierarchy.EligibilityProjectionPoint{
				{EffectiveAt: now, Points: 92, IsEligible: true},
				{EffectiveAt: now.AddDate(0, 0, 7), Points: 82, IsEligible: true},
			},
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/eligibility?subject=player-1", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
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

func TestBanPlayerFromQueueSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		queueBansToList: []hierarchy.QueueBan{
			{
				ID:            1,
				QueueID:       10,
				PlayerID:      20,
				BannedByActor: "support-operator",
				BanReason:     "toxicity",
				IsActive:      true,
				BannedAt:      now,
			},
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/queue-bans", strings.NewReader(`{"queueId":10,"playerId":20,"actor":"support-operator","reason":"toxicity"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestUnbanPlayerFromQueueSuccess(t *testing.T) {
	now := time.Now().UTC()
	reason := "appeal accepted"
	actor := "support-operator"
	store := &fakeHierarchyStore{
		queueBansToList: []hierarchy.QueueBan{
			{
				ID:              1,
				QueueID:         10,
				PlayerID:        20,
				BannedByActor:   "support-operator",
				BanReason:       "toxicity",
				IsActive:        false,
				BannedAt:        now.Add(-time.Hour),
				UnbannedByActor: &actor,
				UnbanReason:     &reason,
				UnbannedAt:      &now,
			},
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/queue-bans/lift", strings.NewReader(`{"queueId":10,"playerId":20,"actor":"support-operator","reason":"appeal accepted"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
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
	req.Header.Set("Authorization", "Bearer local:alice:admin")
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
	req.Header.Set("Authorization", "Bearer local:alice:admin")
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

func TestAdjustPlayerRatingSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		ratingsToList: []hierarchy.PlayerRating{
			{
				ID:            1,
				PlayerID:      20,
				ContextKey:    "scrim-3v3",
				Rating:        1110,
				Uncertainty:   200,
				MatchesPlayed: 25,
				IsActive:      true,
				UpdatedAt:     now,
			},
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{HierarchyStore: store})

	req := httptest.NewRequest(http.MethodPost, "/v1/player-ratings/adjust", strings.NewReader(`{"actorPlayerId":10,"targetPlayerId":20,"contextKey":"scrim-3v3","rating":1110,"uncertainty":200,"matchesPlayed":25,"reason":"manual review"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestAdjustPlayerRatingConflict(t *testing.T) {
	store := &fakeHierarchyStore{
		adjustRatingErr: hierarchy.ErrConflict,
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{HierarchyStore: store})

	req := httptest.NewRequest(http.MethodPost, "/v1/player-ratings/adjust", strings.NewReader(`{"actorPlayerId":10,"targetPlayerId":10,"contextKey":"scrim-3v3","rating":1110,"uncertainty":200,"matchesPlayed":25,"reason":"manual review"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusConflict, rr.Code, rr.Body.String())
	}
}

func TestListRatingAdjustmentsSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		ratingAdjustmentsToList: []hierarchy.RatingAdjustment{
			{
				ID:                    1,
				ActorPlayerID:         10,
				TargetPlayerID:        20,
				ContextKey:            "scrim-3v3",
				PreviousRating:        1000,
				NewRating:             1110,
				PreviousUncertainty:   300,
				NewUncertainty:        200,
				PreviousMatchesPlayed: 20,
				NewMatchesPlayed:      25,
				Reason:                "manual review",
				CreatedAt:             now,
			},
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{HierarchyStore: store})

	req := httptest.NewRequest(http.MethodGet, "/v1/rating-adjustments", nil)
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

func TestOverrideResultSubmissionSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		submissionToReturn: hierarchy.ResultSubmission{
			ID:                1,
			ContextType:       "scrim",
			ContextID:         10,
			SubmittedByTeamID: 20,
			HomeTeamID:        20,
			AwayTeamID:        30,
			WinningTeamID:     30,
			LosingTeamID:      20,
			State:             "ratified",
			PayloadJSON:       []byte(`{"score":"2-3"}`),
			CreatedAt:         now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/result-overrides", strings.NewReader(`{"submissionId":1,"actor":"league-admin","reason":"manual correction","winningTeamId":30,"losingTeamId":20}`))
	req.Header.Set("Authorization", "Bearer local:alice:admin")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}
}

func TestListResultOverridesSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		resultOverridesToList: []hierarchy.ResultOverride{
			{
				ID:                    1,
				SubmissionID:          10,
				Actor:                 "league-admin",
				Reason:                "manual correction",
				PreviousWinningTeamID: 20,
				PreviousLosingTeamID:  30,
				NewWinningTeamID:      30,
				NewLosingTeamID:       20,
				PreviousState:         "pending",
				NewState:              "ratified",
				CreatedAt:             now,
			},
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/result-overrides", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
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
				ParserName:         "kinetic-rl-parser",
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

	req := httptest.NewRequest(http.MethodPost, "/v1/replay-evidence", strings.NewReader(`{"contextType":"scrim","contextId":10,"submittedByTeamId":20,"replayBody":"fake-bytes","parserName":"kinetic-rl-parser","parserVersion":"v0.1.0","parserConfigDigest":"cfg-001","resultSubmissionId":1,"parseOutputJson":{"goals":4}}`))
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
				ParserName:         "kinetic-rl-parser",
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

func TestReportExceptionSuccess(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		exceptionToReturn: hierarchy.ExceptionTicket{
			ID:              1,
			Category:        "scheduling_conflict",
			ContextType:     "match",
			ContextID:       10,
			State:           "open",
			ReasonCode:      "time_unavailable",
			Severity:        3,
			SuggestedAction: "propose_reschedule",
			DetailsJSON:     []byte(`{}`),
			OpenedAt:        now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{HierarchyStore: store})
	req := httptest.NewRequest(http.MethodPost, "/v1/exceptions/report", strings.NewReader(`{"category":"scheduling_conflict","contextType":"match","contextId":10,"reasonCode":"time_unavailable","severity":3,"suggestedAction":"propose_reschedule","detailsJson":{}}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestListExceptionMetricsSuccess(t *testing.T) {
	store := &fakeHierarchyStore{
		exceptionMetricsToReturn: hierarchy.ExceptionMetrics{
			AdminHoursPerWeek:       3.5,
			ManualTouchesPerFixture: 1.2,
			ZeroTouchFixtureRate:    0.65,
			TimeToCloseHoursP50:     8,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{HierarchyStore: store})
	req := httptest.NewRequest(http.MethodGet, "/v1/exception-metrics", nil)
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

	if payload.Service != "kinetic-v3-api" {
		t.Fatalf("unexpected service value %q", payload.Service)
	}
}

// fakeAPITokenStore is a no-op implementation of apitoken.Store for use in tests.
type fakeAPITokenStore struct{}

func (f *fakeAPITokenStore) CreateAPIToken(_ context.Context, _ apitoken.CreateAPITokenInput) (apitoken.APIToken, string, error) {
	return apitoken.APIToken{}, "", nil
}

func (f *fakeAPITokenStore) ListAPITokens(_ context.Context, _ string) ([]apitoken.APIToken, error) {
	return []apitoken.APIToken{}, nil
}

func (f *fakeAPITokenStore) RevokeAPIToken(_ context.Context, _ apitoken.RevokeAPITokenInput) (apitoken.APIToken, error) {
	return apitoken.APIToken{}, nil
}

func (f *fakeAPITokenStore) ValidateAPIToken(_ context.Context, _ string) (apitoken.ValidateAPITokenResult, error) {
	return apitoken.ValidateAPITokenResult{}, nil
}

// --- Theme 1.2: Scope-aware authorization enforcement tests ---

func TestAssignRoleRequiresAuthentication(t *testing.T) {
	store := &fakeHierarchyStore{}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/role-assignments", strings.NewReader(`{"actorPlayerId":1,"targetPlayerId":2,"role":"gm","reason":"test"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}

func TestAssignRoleForbiddenForPlayerRole(t *testing.T) {
	store := &fakeHierarchyStore{}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/role-assignments", strings.NewReader(`{"actorPlayerId":1,"targetPlayerId":2,"role":"gm","reason":"test"}`))
	req.Header.Set("Authorization", "Bearer local:bob:player")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusForbidden, rr.Code, rr.Body.String())
	}
}

func TestRevokeRoleRequiresAuthentication(t *testing.T) {
	store := &fakeHierarchyStore{}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/role-assignments/revoke", strings.NewReader(`{"actorPlayerId":1,"assignmentId":1,"reason":"test"}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}

func TestResultOverrideRequiresAuthentication(t *testing.T) {
	store := &fakeHierarchyStore{}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/result-overrides", strings.NewReader(`{"submissionId":1,"actor":"test","reason":"test","winningTeamId":1,"losingTeamId":2}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}

func TestResultOverrideForbiddenForPlayerRole(t *testing.T) {
	store := &fakeHierarchyStore{}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/result-overrides", strings.NewReader(`{"submissionId":1,"actor":"test","reason":"test","winningTeamId":1,"losingTeamId":2}`))
	req.Header.Set("Authorization", "Bearer local:bob:player")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusForbidden, rr.Code, rr.Body.String())
	}
}

func TestScrimPromotionRequiresAuthentication(t *testing.T) {
	store := &fakeHierarchyStore{}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/scrim-promotions", strings.NewReader(`{"queueId":1}`))
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}

func TestFranchiseManagerCanAssignRole(t *testing.T) {
	now := time.Now().UTC()
	store := &fakeHierarchyStore{
		roleAssignmentToReturn: hierarchy.RoleAssignment{
			ID:                      1,
			PlayerID:                22,
			Role:                    "captain",
			AssignedByActorPlayerID: 11,
			AssignReason:            "team staffing",
			IsActive:                true,
			AssignedAt:              now,
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{
		HierarchyStore: store,
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/role-assignments", strings.NewReader(`{"actorPlayerId":11,"targetPlayerId":22,"role":"captain","reason":"team staffing"}`))
	req.Header.Set("Authorization", "Bearer local:alice:fm")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
}

func TestCreateMyPlayerUsesAuthenticatedUserAndGame(t *testing.T) {
	store := &fakeHierarchyStore{
		userToReturn: hierarchy.User{ID: 17, Subject: "jake", DisplayName: "Jake"},
		userPlayerToReturn: hierarchy.UserPlayer{
			UserID: 17, PlayerID: 42, GameID: 3,
			Player: hierarchy.Player{ID: 42, DisplayName: "Rocket League Jake", Slug: "rocket-league-jake"},
			Game:   hierarchy.Game{ID: 3, Name: "Rocket League", Slug: "rocket-league"},
		},
	}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{HierarchyStore: store})
	req := httptest.NewRequest(http.MethodPost, "/v1/me/players", strings.NewReader(`{"gameId":3,"displayName":"Rocket League Jake","slug":"rocket-league-jake"}`))
	req.Header.Set("Authorization", "Bearer local:jake:player")
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}
	if store.createUserPlayerInput.UserID != 17 || store.createUserPlayerInput.GameID != 3 {
		t.Fatalf("expected authenticated user 17 and game 3, got %#v", store.createUserPlayerInput)
	}
}

func TestMyPlayersRequiresAuthentication(t *testing.T) {
	store := &fakeHierarchyStore{}
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{HierarchyStore: store})
	req := httptest.NewRequest(http.MethodGet, "/v1/me/players", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnauthorized, rr.Code, rr.Body.String())
	}
}

func TestAuthProvidersEndpoint(t *testing.T) {
	t.Run("discord enabled when credentials are set", func(t *testing.T) {
		cfg := config.Config{
			Port:                "8080",
			LogLevel:            "info",
			DiscordClientID:     "test-client-id",
			DiscordClientSecret: "test-client-secret",
		}
		srv := New(cfg, slog.Default(), Dependencies{})

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/providers", nil)
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var providers []map[string]any
		if err := json.NewDecoder(rr.Body).Decode(&providers); err != nil {
			t.Fatalf("failed to decode providers response: %v", err)
		}

		if len(providers) == 0 {
			t.Fatal("expected at least one provider in response")
		}

		var found bool
		for _, p := range providers {
			if p["id"] == "discord" {
				found = true
				if p["enabled"] != true {
					t.Fatalf("expected discord to be enabled, got %v", p["enabled"])
				}
				if p["name"] != "Discord" {
					t.Fatalf("expected discord name to be 'Discord', got %v", p["name"])
				}
			}
		}
		if !found {
			t.Fatal("expected discord entry in providers list")
		}
	})

	t.Run("discord disabled when credentials are absent", func(t *testing.T) {
		cfg := config.Config{
			Port:     "8080",
			LogLevel: "info",
		}
		srv := New(cfg, slog.Default(), Dependencies{})

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/providers", nil)
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
		}

		var providers []map[string]any
		if err := json.NewDecoder(rr.Body).Decode(&providers); err != nil {
			t.Fatalf("failed to decode providers response: %v", err)
		}

		for _, p := range providers {
			if p["id"] == "discord" {
				if p["enabled"] != false {
					t.Fatalf("expected discord to be disabled when credentials absent, got %v", p["enabled"])
				}
				return
			}
		}
		t.Fatal("expected discord entry in providers list")
	})

	t.Run("rejects non-GET methods", func(t *testing.T) {
		srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

		req := httptest.NewRequest(http.MethodPost, "/v1/auth/providers", nil)
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
		}
	})
}

func TestAuthLoginDiscordRedirectsWhenConfigured(t *testing.T) {
	cfg := config.Config{
		Port:                "8080",
		LogLevel:            "info",
		WebBaseURL:          "http://localhost:5173",
		DiscordClientID:     "test-client-id",
		DiscordClientSecret: "test-client-secret",
	}
	srv := New(cfg, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/login?provider=discord", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusFound {
		t.Fatalf("expected redirect %d, got %d body=%s", http.StatusFound, rr.Code, rr.Body.String())
	}

	location := rr.Result().Header.Get("Location")
	if !strings.Contains(location, "discord.com/api/oauth2/authorize") {
		t.Fatalf("expected redirect to discord authorize URL, got %s", location)
	}
	if !strings.Contains(location, "client_id=test-client-id") {
		t.Fatalf("expected client_id in discord URL, got %s", location)
	}
	if !strings.Contains(location, "state=") {
		t.Fatalf("expected state parameter in discord URL, got %s", location)
	}

	// Verify state cookie was set.
	var stateCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "oauth_state" {
			stateCookie = c
			break
		}
	}
	if stateCookie == nil {
		t.Fatal("expected oauth_state cookie to be set")
	}
	if stateCookie.MaxAge != 600 {
		t.Fatalf("expected oauth_state cookie max-age 600, got %d", stateCookie.MaxAge)
	}
}

func TestAuthLoginDiscordReturns501WhenNotConfigured(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/login?provider=discord", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("expected status %d, got %d", http.StatusNotImplemented, rr.Code)
	}
}

func TestAuthLoginUnknownProviderReturns400(t *testing.T) {
	srv := New(config.Config{Port: "8080", LogLevel: "info"}, slog.Default(), Dependencies{})

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/login?provider=github", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestAuthCallbackDiscordMissingStateReturns400(t *testing.T) {
	srv := New(config.Config{
		Port:                "8080",
		LogLevel:            "info",
		DiscordClientID:     "test-client-id",
		DiscordClientSecret: "test-client-secret",
	}, slog.Default(), Dependencies{})

	// Call callback with a code but no state cookie set (state validation should fail).
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/callback?provider=discord&code=somecode&state=mismatch", nil)
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}
