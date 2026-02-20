package hierarchy

import (
	"encoding/json"
	"time"
)

type League struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

type Franchise struct {
	ID        int64     `json:"id"`
	LeagueID  int64     `json:"leagueId"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

type Club struct {
	ID          int64     `json:"id"`
	FranchiseID int64     `json:"franchiseId"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Team struct {
	ID        int64     `json:"id"`
	ClubID    int64     `json:"clubId"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

type Player struct {
	ID          int64     `json:"id"`
	DisplayName string    `json:"displayName"`
	Slug        string    `json:"slug"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
}

type RosterMembership struct {
	ID        int64     `json:"id"`
	PlayerID  int64     `json:"playerId"`
	TeamID    int64     `json:"teamId"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

type Queue struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

type QueueEntry struct {
	ID        int64      `json:"id"`
	QueueID   int64      `json:"queueId"`
	TeamID    int64      `json:"teamId"`
	IsActive  bool       `json:"isActive"`
	Stage     int32      `json:"stage"`
	CreatedAt time.Time  `json:"createdAt"`
	StageAt   time.Time  `json:"stageAt"`
	LeftAt    *time.Time `json:"leftAt,omitempty"`
}

type QueueBan struct {
	ID              int64      `json:"id"`
	QueueID         int64      `json:"queueId"`
	PlayerID        int64      `json:"playerId"`
	BannedByActor   string     `json:"bannedByActor"`
	BanReason       string     `json:"banReason"`
	IsActive        bool       `json:"isActive"`
	BannedAt        time.Time  `json:"bannedAt"`
	UnbannedByActor *string    `json:"unbannedByActor,omitempty"`
	UnbanReason     *string    `json:"unbanReason,omitempty"`
	UnbannedAt      *time.Time `json:"unbannedAt,omitempty"`
}

type PlatformAccountLink struct {
	ID                  int64      `json:"id"`
	Subject             string     `json:"subject"`
	Provider            string     `json:"provider"`
	ProviderAccountID   string     `json:"providerAccountId"`
	ProviderAccountName string     `json:"providerAccountName"`
	IsActive            bool       `json:"isActive"`
	LinkedAt            time.Time  `json:"linkedAt"`
	UnlinkedAt          *time.Time `json:"unlinkedAt,omitempty"`
}

type EligibilityProjectionPoint struct {
	EffectiveAt time.Time `json:"effectiveAt"`
	Points      int32     `json:"points"`
	IsEligible  bool      `json:"isEligible"`
}

type EligibilityStatus struct {
	Subject         string                       `json:"subject"`
	Points          int32                        `json:"points"`
	ThresholdPoints int32                        `json:"thresholdPoints"`
	DecayPerWeek    int32                        `json:"decayPerWeek"`
	EligibleUntil   time.Time                    `json:"eligibleUntil"`
	EvaluatedAt     time.Time                    `json:"evaluatedAt"`
	Projection      []EligibilityProjectionPoint `json:"projection"`
}

type Scrim struct {
	ID         int64      `json:"id"`
	QueueID    int64      `json:"queueId"`
	HomeTeamID int64      `json:"homeTeamId"`
	AwayTeamID int64      `json:"awayTeamId"`
	State      string     `json:"state"`
	CreatedAt  time.Time  `json:"createdAt"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	EndedAt    *time.Time `json:"endedAt,omitempty"`
}

type PlayerRating struct {
	ID             int64      `json:"id"`
	PlayerID       int64      `json:"playerId"`
	ContextKey     string     `json:"contextKey"`
	Rating         int32      `json:"rating"`
	Uncertainty    int32      `json:"uncertainty"`
	MatchesPlayed  int32      `json:"matchesPlayed"`
	LastCompetedAt *time.Time `json:"lastCompetedAt,omitempty"`
	IsActive       bool       `json:"isActive"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type RatingAdjustment struct {
	ID                    int64     `json:"id"`
	ActorPlayerID         int64     `json:"actorPlayerId"`
	TargetPlayerID        int64     `json:"targetPlayerId"`
	ContextKey            string    `json:"contextKey"`
	PreviousRating        int32     `json:"previousRating"`
	NewRating             int32     `json:"newRating"`
	PreviousUncertainty   int32     `json:"previousUncertainty"`
	NewUncertainty        int32     `json:"newUncertainty"`
	PreviousMatchesPlayed int32     `json:"previousMatchesPlayed"`
	NewMatchesPlayed      int32     `json:"newMatchesPlayed"`
	Reason                string    `json:"reason"`
	CreatedAt             time.Time `json:"createdAt"`
}

type MatchmakingDecision struct {
	ID               int64     `json:"id"`
	ScrimID          int64     `json:"scrimId"`
	QueueID          int64     `json:"queueId"`
	QueueWaitSeconds int32     `json:"queueWaitSeconds"`
	WaitSkewSeconds  int32     `json:"waitSkewSeconds"`
	ExpansionStage   int32     `json:"expansionStage"`
	RatingSpread     int32     `json:"ratingSpread"`
	HomeTeamRating   int32     `json:"homeTeamRating"`
	AwayTeamRating   int32     `json:"awayTeamRating"`
	CrossGroup       bool      `json:"crossGroup"`
	OrderingStrategy string    `json:"orderingStrategy"`
	CreatedAt        time.Time `json:"createdAt"`
}

type ResultSubmission struct {
	ID                int64           `json:"id"`
	ContextType       string          `json:"contextType"`
	ContextID         int64           `json:"contextId"`
	SubmittedByTeamID int64           `json:"submittedByTeamId"`
	HomeTeamID        int64           `json:"homeTeamId"`
	AwayTeamID        int64           `json:"awayTeamId"`
	WinningTeamID     int64           `json:"winningTeamId"`
	LosingTeamID      int64           `json:"losingTeamId"`
	State             string          `json:"state"`
	PayloadJSON       json.RawMessage `json:"payloadJson"`
	HomeRatifiedAt    *time.Time      `json:"homeRatifiedAt,omitempty"`
	AwayRatifiedAt    *time.Time      `json:"awayRatifiedAt,omitempty"`
	RejectedByTeamID  *int64          `json:"rejectedByTeamId,omitempty"`
	RejectionReason   *string         `json:"rejectionReason,omitempty"`
	RejectedAt        *time.Time      `json:"rejectedAt,omitempty"`
	CreatedAt         time.Time       `json:"createdAt"`
}

type ResultOverride struct {
	ID                    int64     `json:"id"`
	SubmissionID          int64     `json:"submissionId"`
	Actor                 string    `json:"actor"`
	Reason                string    `json:"reason"`
	PreviousWinningTeamID int64     `json:"previousWinningTeamId"`
	PreviousLosingTeamID  int64     `json:"previousLosingTeamId"`
	NewWinningTeamID      int64     `json:"newWinningTeamId"`
	NewLosingTeamID       int64     `json:"newLosingTeamId"`
	PreviousState         string    `json:"previousState"`
	NewState              string    `json:"newState"`
	CreatedAt             time.Time `json:"createdAt"`
}

type ReplayEvidence struct {
	ID                int64     `json:"id"`
	ContextType       string    `json:"contextType"`
	ContextID         int64     `json:"contextId"`
	SubmittedByTeamID int64     `json:"submittedByTeamId"`
	ReplaySHA256      string    `json:"replaySha256"`
	ContentSizeBytes  int64     `json:"contentSizeBytes"`
	StorageRef        string    `json:"storageRef"`
	State             string    `json:"state"`
	CreatedAt         time.Time `json:"createdAt"`
}

type ReplayParseRun struct {
	ID                 int64           `json:"id"`
	ReplayEvidenceID   int64           `json:"replayEvidenceId"`
	ParserName         string          `json:"parserName"`
	ParserVersion      string          `json:"parserVersion"`
	ParserConfigDigest string          `json:"parserConfigDigest"`
	Status             string          `json:"status"`
	OutputJSON         json.RawMessage `json:"outputJson"`
	CreatedAt          time.Time       `json:"createdAt"`
}

type ResultSubmissionReplayLink struct {
	ID                 int64     `json:"id"`
	ResultSubmissionID int64     `json:"resultSubmissionId"`
	ReplayEvidenceID   int64     `json:"replayEvidenceId"`
	LinkedByTeamID     int64     `json:"linkedByTeamId"`
	CreatedAt          time.Time `json:"createdAt"`
}

type ExceptionTicket struct {
	ID               int64           `json:"id"`
	Category         string          `json:"category"`
	ContextType      string          `json:"contextType"`
	ContextID        int64           `json:"contextId"`
	ReportedByTeamID *int64          `json:"reportedByTeamId,omitempty"`
	State            string          `json:"state"`
	ReasonCode       string          `json:"reasonCode"`
	Severity         int32           `json:"severity"`
	SuggestedAction  string          `json:"suggestedAction"`
	DetailsJSON      json.RawMessage `json:"detailsJson"`
	ResolutionCode   *string         `json:"resolutionCode,omitempty"`
	OpenedAt         time.Time       `json:"openedAt"`
	TriagedAt        *time.Time      `json:"triagedAt,omitempty"`
	ResolvedAt       *time.Time      `json:"resolvedAt,omitempty"`
}

type ExceptionAction struct {
	ID           int64     `json:"id"`
	TicketID     int64     `json:"ticketId"`
	ActionType   string    `json:"actionType"`
	Actor        string    `json:"actor"`
	Automated    bool      `json:"automated"`
	Notes        string    `json:"notes"`
	MinutesSpent int32     `json:"minutesSpent"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ExceptionMetrics struct {
	AdminHoursPerWeek       float64 `json:"adminHoursPerWeek"`
	ManualTouchesPerFixture float64 `json:"manualTouchesPerFixture"`
	ZeroTouchFixtureRate    float64 `json:"zeroTouchFixtureRate"`
	TimeToCloseHoursP50     float64 `json:"timeToCloseHoursP50"`
}

type ExceptionAutomationResult struct {
	Ticket       ExceptionTicket `json:"ticket"`
	AutoResolved bool            `json:"autoResolved"`
}

type Season struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

type ScheduleGroup struct {
	ID        int64     `json:"id"`
	SeasonID  int64     `json:"seasonId"`
	Name      string    `json:"name"`
	Sequence  int32     `json:"sequence"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

type Fixture struct {
	ID              int64     `json:"id"`
	ScheduleGroupID int64     `json:"scheduleGroupId"`
	HomeClubID      int64     `json:"homeClubId"`
	AwayClubID      int64     `json:"awayClubId"`
	IsActive        bool      `json:"isActive"`
	CreatedAt       time.Time `json:"createdAt"`
}

type Match struct {
	ID                 int64      `json:"id"`
	FixtureID          int64      `json:"fixtureId"`
	HomeTeamID         int64      `json:"homeTeamId"`
	AwayTeamID         int64      `json:"awayTeamId"`
	State              string     `json:"state"`
	ScheduledFor       *time.Time `json:"scheduledFor,omitempty"`
	HomeTimeRatifiedAt *time.Time `json:"homeTimeRatifiedAt,omitempty"`
	AwayTimeRatifiedAt *time.Time `json:"awayTimeRatifiedAt,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
}

type CreateLeagueInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type CreateFranchiseInput struct {
	LeagueID int64  `json:"leagueId"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
}

type CreateClubInput struct {
	FranchiseID int64  `json:"franchiseId"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
}

type CreateTeamInput struct {
	ClubID int64  `json:"clubId"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
}

type CreatePlayerInput struct {
	DisplayName string `json:"displayName"`
	Slug        string `json:"slug"`
}

type CreateRosterMembershipInput struct {
	PlayerID int64 `json:"playerId"`
	TeamID   int64 `json:"teamId"`
}

type CreateQueueInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type EnqueueTeamInput struct {
	QueueID int64 `json:"queueId"`
	TeamID  int64 `json:"teamId"`
}

type LeaveQueueInput struct {
	QueueID int64 `json:"queueId"`
	TeamID  int64 `json:"teamId"`
}

type BanPlayerFromQueueInput struct {
	QueueID  int64  `json:"queueId"`
	PlayerID int64  `json:"playerId"`
	Actor    string `json:"actor"`
	Reason   string `json:"reason"`
}

type UnbanPlayerFromQueueInput struct {
	QueueID  int64  `json:"queueId"`
	PlayerID int64  `json:"playerId"`
	Actor    string `json:"actor"`
	Reason   string `json:"reason"`
}

type LinkPlatformAccountInput struct {
	Subject             string `json:"subject"`
	Provider            string `json:"provider"`
	ProviderAccountID   string `json:"providerAccountId"`
	ProviderAccountName string `json:"providerAccountName"`
}

type UnlinkPlatformAccountInput struct {
	Subject           string `json:"subject"`
	Provider          string `json:"provider"`
	ProviderAccountID string `json:"providerAccountId"`
}

type GetEligibilityInput struct {
	Subject string `json:"subject"`
}

type AdvanceQueueEntryStageInput struct {
	QueueID int64 `json:"queueId"`
	TeamID  int64 `json:"teamId"`
	Stage   int32 `json:"stage"`
}

type CreateScrimInput struct {
	QueueID    int64  `json:"queueId"`
	HomeTeamID int64  `json:"homeTeamId"`
	AwayTeamID int64  `json:"awayTeamId"`
	State      string `json:"state"`
}

type PromoteQueueToScrimInput struct {
	QueueID int64 `json:"queueId"`
}

type UpdateScrimStateInput struct {
	ScrimID int64  `json:"scrimId"`
	State   string `json:"state"`
}

type ProcessQueuePromotionsInput struct {
	QueueID int64 `json:"queueId"`
}

type ProcessQueuePromotionsResult struct {
	ProcessedQueues   int32 `json:"processedQueues"`
	PromotionsCreated int32 `json:"promotionsCreated"`
	Conflicts         int32 `json:"conflicts"`
}

type AdjustPlayerRatingInput struct {
	ActorPlayerID  int64  `json:"actorPlayerId"`
	TargetPlayerID int64  `json:"targetPlayerId"`
	ContextKey     string `json:"contextKey"`
	Rating         int32  `json:"rating"`
	Uncertainty    int32  `json:"uncertainty"`
	MatchesPlayed  int32  `json:"matchesPlayed"`
	Reason         string `json:"reason"`
}

type PromotionProcessingRun struct {
	ID                int64     `json:"id"`
	QueueID           *int64    `json:"queueId,omitempty"`
	ProcessedQueues   int32     `json:"processedQueues"`
	PromotionsCreated int32     `json:"promotionsCreated"`
	Conflicts         int32     `json:"conflicts"`
	DurationMs        int32     `json:"durationMs"`
	CreatedAt         time.Time `json:"createdAt"`
}

type CreateResultSubmissionInput struct {
	ContextType       string          `json:"contextType"`
	ContextID         int64           `json:"contextId"`
	SubmittedByTeamID int64           `json:"submittedByTeamId"`
	WinningTeamID     int64           `json:"winningTeamId"`
	LosingTeamID      int64           `json:"losingTeamId"`
	PayloadJSON       json.RawMessage `json:"payloadJson"`
}

type OverrideResultSubmissionInput struct {
	SubmissionID  int64  `json:"submissionId"`
	Actor         string `json:"actor"`
	Reason        string `json:"reason"`
	WinningTeamID int64  `json:"winningTeamId"`
	LosingTeamID  int64  `json:"losingTeamId"`
}

type RatifyResultSubmissionInput struct {
	SubmissionID int64 `json:"submissionId"`
	TeamID       int64 `json:"teamId"`
}

type RejectResultSubmissionInput struct {
	SubmissionID int64  `json:"submissionId"`
	TeamID       int64  `json:"teamId"`
	Reason       string `json:"reason"`
}

type IngestReplayEvidenceInput struct {
	ContextType        string          `json:"contextType"`
	ContextID          int64           `json:"contextId"`
	SubmittedByTeamID  int64           `json:"submittedByTeamId"`
	ReplayBody         string          `json:"replayBody"`
	ParserName         string          `json:"parserName"`
	ParserVersion      string          `json:"parserVersion"`
	ParserConfigDigest string          `json:"parserConfigDigest"`
	ParseOutputJSON    json.RawMessage `json:"parseOutputJson"`
	ResultSubmissionID *int64          `json:"resultSubmissionId,omitempty"`
}

type ReplayIngestionResult struct {
	Evidence           ReplayEvidence `json:"evidence"`
	ParseRun           ReplayParseRun `json:"parseRun"`
	Duplicate          bool           `json:"duplicate"`
	LinkedSubmissionID *int64         `json:"linkedSubmissionId,omitempty"`
}

type ReportExceptionInput struct {
	Category         string          `json:"category"`
	ContextType      string          `json:"contextType"`
	ContextID        int64           `json:"contextId"`
	ReportedByTeamID *int64          `json:"reportedByTeamId,omitempty"`
	ReasonCode       string          `json:"reasonCode"`
	Severity         int32           `json:"severity"`
	SuggestedAction  string          `json:"suggestedAction"`
	DetailsJSON      json.RawMessage `json:"detailsJson"`
}

type TriageExceptionInput struct {
	TicketID        int64  `json:"ticketId"`
	Actor           string `json:"actor"`
	ReasonCode      string `json:"reasonCode"`
	Severity        int32  `json:"severity"`
	SuggestedAction string `json:"suggestedAction"`
	MinutesSpent    int32  `json:"minutesSpent"`
}

type ResolveExceptionInput struct {
	TicketID       int64  `json:"ticketId"`
	Actor          string `json:"actor"`
	ResolutionCode string `json:"resolutionCode"`
	Notes          string `json:"notes"`
	Automated      bool   `json:"automated"`
	MinutesSpent   int32  `json:"minutesSpent"`
}

type EvaluateSchedulingExceptionInput struct {
	MatchID       int64  `json:"matchId"`
	ConflictCode  string `json:"conflictCode"`
	HomeConfirmed bool   `json:"homeConfirmed"`
	AwayConfirmed bool   `json:"awayConfirmed"`
	Actor         string `json:"actor"`
}

type EvaluateNoShowExceptionInput struct {
	MatchID       int64  `json:"matchId"`
	HomeCheckedIn bool   `json:"homeCheckedIn"`
	AwayCheckedIn bool   `json:"awayCheckedIn"`
	GraceMinutes  int32  `json:"graceMinutes"`
	Actor         string `json:"actor"`
}

type EvaluateReplayDisputeExceptionInput struct {
	ResultSubmissionID int64  `json:"resultSubmissionId"`
	ParseStatus        string `json:"parseStatus"`
	IdentityStatus     string `json:"identityStatus"`
	DisputeRaised      bool   `json:"disputeRaised"`
	Actor              string `json:"actor"`
}

type CreateSeasonInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type CreateScheduleGroupInput struct {
	SeasonID int64  `json:"seasonId"`
	Name     string `json:"name"`
	Sequence int32  `json:"sequence"`
}

type CreateFixtureInput struct {
	ScheduleGroupID int64 `json:"scheduleGroupId"`
	HomeClubID      int64 `json:"homeClubId"`
	AwayClubID      int64 `json:"awayClubId"`
}

type CreateMatchInput struct {
	FixtureID          int64      `json:"fixtureId"`
	HomeTeamID         int64      `json:"homeTeamId"`
	AwayTeamID         int64      `json:"awayTeamId"`
	State              string     `json:"state"`
	ScheduledFor       *time.Time `json:"scheduledFor,omitempty"`
	HomeTimeRatifiedAt *time.Time `json:"homeTimeRatifiedAt,omitempty"`
	AwayTimeRatifiedAt *time.Time `json:"awayTimeRatifiedAt,omitempty"`
}
