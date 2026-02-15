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
