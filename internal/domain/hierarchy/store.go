package hierarchy

import "context"

type Store interface {
	CreateLeague(ctx context.Context, input CreateLeagueInput) (League, error)
	ListLeagues(ctx context.Context) ([]League, error)
	CreateFranchise(ctx context.Context, input CreateFranchiseInput) (Franchise, error)
	ListFranchises(ctx context.Context) ([]Franchise, error)
	CreateClub(ctx context.Context, input CreateClubInput) (Club, error)
	ListClubs(ctx context.Context) ([]Club, error)
}
