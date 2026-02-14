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
