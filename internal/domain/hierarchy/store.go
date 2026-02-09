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
}
