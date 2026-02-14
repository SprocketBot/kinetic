package hierarchy

import "time"

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
	CreatedAt time.Time  `json:"createdAt"`
	LeftAt    *time.Time `json:"leftAt,omitempty"`
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
