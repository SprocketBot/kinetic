package hierarchy

import "context"

type Store interface {
	CreateLeague(ctx context.Context, input CreateLeagueInput) (League, error)
	ListLeagues(ctx context.Context) ([]League, error)
	CreateFranchise(ctx context.Context, input CreateFranchiseInput) (Franchise, error)
	ListFranchises(ctx context.Context) ([]Franchise, error)
	CreateClub(ctx context.Context, input CreateClubInput) (Club, error)
	ListClubs(ctx context.Context) ([]Club, error)
	CreateTeam(ctx context.Context, input CreateTeamInput) (Team, error)
	ListTeams(ctx context.Context) ([]Team, error)
	CreatePlayer(ctx context.Context, input CreatePlayerInput) (Player, error)
	ListPlayers(ctx context.Context) ([]Player, error)
	CreateRosterMembership(ctx context.Context, input CreateRosterMembershipInput) (RosterMembership, error)
	ListRosterMemberships(ctx context.Context) ([]RosterMembership, error)
	CreateQueue(ctx context.Context, input CreateQueueInput) (Queue, error)
	ListQueues(ctx context.Context) ([]Queue, error)
	EnqueueTeam(ctx context.Context, input EnqueueTeamInput) (QueueEntry, error)
	LeaveQueue(ctx context.Context, input LeaveQueueInput) (QueueEntry, error)
	ListActiveQueueEntries(ctx context.Context) ([]QueueEntry, error)
	CreateSeason(ctx context.Context, input CreateSeasonInput) (Season, error)
	ListSeasons(ctx context.Context) ([]Season, error)
	CreateScheduleGroup(ctx context.Context, input CreateScheduleGroupInput) (ScheduleGroup, error)
	ListScheduleGroups(ctx context.Context) ([]ScheduleGroup, error)
	CreateFixture(ctx context.Context, input CreateFixtureInput) (Fixture, error)
	ListFixtures(ctx context.Context) ([]Fixture, error)
	CreateMatch(ctx context.Context, input CreateMatchInput) (Match, error)
	ListMatches(ctx context.Context) ([]Match, error)
}
