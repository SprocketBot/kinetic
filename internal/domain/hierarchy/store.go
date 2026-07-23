package hierarchy

import "context"

type LeagueStore interface {
	CreateLeague(ctx context.Context, input CreateLeagueInput) (League, error)
	ListLeagues(ctx context.Context) ([]League, error)
	CreateFranchise(ctx context.Context, input CreateFranchiseInput) (Franchise, error)
	ListFranchises(ctx context.Context) ([]Franchise, error)
	CreateClub(ctx context.Context, input CreateClubInput) (Club, error)
	ListClubs(ctx context.Context) ([]Club, error)
	CreateTeam(ctx context.Context, input CreateTeamInput) (Team, error)
	ListTeams(ctx context.Context) ([]Team, error)
}

type PlayerStore interface {
	CreatePlayer(ctx context.Context, input CreatePlayerInput) (Player, error)
	ListPlayers(ctx context.Context) ([]Player, error)
	SetPlayerActive(ctx context.Context, input SetPlayerActiveInput) (Player, error)
}

type IdentityStore interface {
	UpsertUser(ctx context.Context, input UpsertUserInput) (User, error)
	CreateGame(ctx context.Context, input CreateGameInput) (Game, error)
	ListGames(ctx context.Context) ([]Game, error)
	CreateUserPlayer(ctx context.Context, input CreateUserPlayerInput) (UserPlayer, error)
	ListUserPlayers(ctx context.Context, userID int64) ([]UserPlayer, error)
	UserOwnsPlayer(ctx context.Context, userID, playerID int64) (bool, error)
	GetUserPlayerIDForGame(ctx context.Context, userID, gameID int64) (int64, error)
}

type RosterStore interface {
	CreateRosterMembership(ctx context.Context, input CreateRosterMembershipInput) (RosterMembership, error)
	ListRosterMemberships(ctx context.Context) ([]RosterMembership, error)
	// Me-scoped helper: returns nil (not an error) when the player simply has no active record.
	GetActiveRosterMembershipByPlayerID(ctx context.Context, playerID int64) (*RosterMembership, error)
}

type RoleStore interface {
	AssignRole(ctx context.Context, input AssignRoleInput) (RoleAssignment, error)
	RevokeRole(ctx context.Context, input RevokeRoleInput) (RoleAssignment, error)
	ListRoleAssignments(ctx context.Context) ([]RoleAssignment, error)
	ResolveScopedRoles(ctx context.Context, input ResolveScopedRolesInput) ([]string, error)
	ResolveRoleScope(ctx context.Context, role string, franchiseID, clubID, teamID *int64) (HierarchyScope, error)
	ResolveAssignmentScope(ctx context.Context, assignmentID int64) (HierarchyScope, error)
}

type QueueStore interface {
	CreateQueue(ctx context.Context, input CreateQueueInput) (Queue, error)
	ListQueues(ctx context.Context) ([]Queue, error)
	EnqueueTeam(ctx context.Context, input EnqueueTeamInput) (QueueEntry, error)
	LeaveQueue(ctx context.Context, input LeaveQueueInput) (QueueEntry, error)
	BanPlayerFromQueue(ctx context.Context, input BanPlayerFromQueueInput) (QueueBan, error)
	UnbanPlayerFromQueue(ctx context.Context, input UnbanPlayerFromQueueInput) (QueueBan, error)
	ListQueueBans(ctx context.Context) ([]QueueBan, error)
	AdvanceQueueEntryStage(ctx context.Context, input AdvanceQueueEntryStageInput) (QueueEntry, error)
	ListActiveQueueEntries(ctx context.Context) ([]QueueEntry, error)
	// Me-scoped helpers: return nil or an empty list when the player simply has no active record.
	ListActiveQueueBansByPlayerID(ctx context.Context, playerID int64) ([]QueueBan, error)
	GetActiveQueueEntryByPlayerID(ctx context.Context, playerID int64) (*QueueEntry, error)
}

type PlatformAccountStore interface {
	LinkPlatformAccount(ctx context.Context, input LinkPlatformAccountInput) (PlatformAccountLink, error)
	UnlinkPlatformAccount(ctx context.Context, input UnlinkPlatformAccountInput) (PlatformAccountLink, error)
	ListPlatformAccountLinks(ctx context.Context, subject string) ([]PlatformAccountLink, error)
	ListPlatformAccountLinksByPlayerID(ctx context.Context, playerID int64) ([]PlatformAccountLink, error)
}

type EligibilityStore interface {
	GetEligibilityStatus(ctx context.Context, subject string) (EligibilityStatus, error)
}

type ScrimStore interface {
	CreateScrim(ctx context.Context, input CreateScrimInput) (Scrim, error)
	UpdateScrimState(ctx context.Context, input UpdateScrimStateInput) (Scrim, error)
	ListScrims(ctx context.Context) ([]Scrim, error)
	PromoteQueueToScrim(ctx context.Context, input PromoteQueueToScrimInput) (Scrim, error)
	ProcessQueuePromotions(ctx context.Context, input ProcessQueuePromotionsInput) (ProcessQueuePromotionsResult, error)
	ListPromotionProcessingRuns(ctx context.Context) ([]PromotionProcessingRun, error)
	GetScrim(ctx context.Context, scrimID int64) (Scrim, error)
	CheckInScrim(ctx context.Context, input CheckInScrimInput) (Scrim, error)
	// ExecutePopTimeout cancels a popped scrim that has timed out, applies queue bans to non-checking-in
	// teams' players, and writes player notifications. It is a no-op if the scrim is no longer in popped state.
	ExecutePopTimeout(ctx context.Context, input ExecutePopTimeoutInput) error
	GetScrimMetrics(ctx context.Context) (ScrimMetrics, error)
	// Me-scoped helper: returns nil (not an error) when the player simply has no active scrim.
	GetActiveScrimByPlayerID(ctx context.Context, playerID int64) (*Scrim, error)
}

type RatingStore interface {
	ListPlayerRatings(ctx context.Context) ([]PlayerRating, error)
	AdjustPlayerRating(ctx context.Context, input AdjustPlayerRatingInput) (PlayerRating, error)
	ListRatingAdjustments(ctx context.Context) ([]RatingAdjustment, error)
	ListMatchmakingDecisions(ctx context.Context) ([]MatchmakingDecision, error)
}

type ResultStore interface {
	CreateResultSubmission(ctx context.Context, input CreateResultSubmissionInput) (ResultSubmission, error)
	ListResultSubmissions(ctx context.Context) ([]ResultSubmission, error)
	OverrideResultSubmission(ctx context.Context, input OverrideResultSubmissionInput) (ResultSubmission, error)
	ListResultOverrides(ctx context.Context) ([]ResultOverride, error)
	RatifyResultSubmission(ctx context.Context, input RatifyResultSubmissionInput) (ResultSubmission, error)
	RejectResultSubmission(ctx context.Context, input RejectResultSubmissionInput) (ResultSubmission, error)
	GetResultSubmission(ctx context.Context, submissionID int64) (ResultSubmission, error)
	ListResultSubmissionsFiltered(ctx context.Context, input ListResultSubmissionsInput) ([]ResultSubmission, error)
	ResetResultSubmission(ctx context.Context, input ResetResultSubmissionInput) (ResultSubmission, error)
}

type ReplayStore interface {
	IngestReplayEvidence(ctx context.Context, input IngestReplayEvidenceInput) (ReplayIngestionResult, error)
	ListReplayEvidence(ctx context.Context) ([]ReplayEvidence, error)
	ListReplayParseRuns(ctx context.Context) ([]ReplayParseRun, error)
	ListResultSubmissionReplayLinks(ctx context.Context) ([]ResultSubmissionReplayLink, error)
	// TriggerReplayParse launches a background stub parse of the given replay evidence,
	// creating round and player stat line records linked to the associated result submission.
	TriggerReplayParse(ctx context.Context, evidenceID, contextID int64, contextType string) error
}

type ExceptionStore interface {
	ReportException(ctx context.Context, input ReportExceptionInput) (ExceptionTicket, error)
	ListOperatorInbox(ctx context.Context) ([]ExceptionTicket, error)
	TriageException(ctx context.Context, input TriageExceptionInput) (ExceptionTicket, error)
	ResolveException(ctx context.Context, input ResolveExceptionInput) (ExceptionTicket, error)
	ListExceptionActions(ctx context.Context) ([]ExceptionAction, error)
	GetExceptionMetrics(ctx context.Context) (ExceptionMetrics, error)
	EvaluateSchedulingException(ctx context.Context, input EvaluateSchedulingExceptionInput) (ExceptionAutomationResult, error)
	EvaluateNoShowException(ctx context.Context, input EvaluateNoShowExceptionInput) (ExceptionAutomationResult, error)
	EvaluateReplayDisputeException(ctx context.Context, input EvaluateReplayDisputeExceptionInput) (ExceptionAutomationResult, error)
}

type SchedulingStore interface {
	CreateSeason(ctx context.Context, input CreateSeasonInput) (Season, error)
	ListSeasons(ctx context.Context) ([]Season, error)
	CreateScheduleGroup(ctx context.Context, input CreateScheduleGroupInput) (ScheduleGroup, error)
	ListScheduleGroups(ctx context.Context) ([]ScheduleGroup, error)
	CreateFixture(ctx context.Context, input CreateFixtureInput) (Fixture, error)
	ListFixtures(ctx context.Context) ([]Fixture, error)
	GetFixture(ctx context.Context, fixtureID int64) (Fixture, error)
	CreateMatch(ctx context.Context, input CreateMatchInput) (Match, error)
	ListMatches(ctx context.Context) ([]Match, error)
}

type Store interface {
	LeagueStore
	PlayerStore
	IdentityStore
	RosterStore
	RoleStore
	QueueStore
	PlatformAccountStore
	EligibilityStore
	ScrimStore
	RatingStore
	ResultStore
	ReplayStore
	ExceptionStore
	SchedulingStore
}
