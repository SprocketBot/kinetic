package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sprocketbot/sprocket-v3/internal/domain/hierarchy"
)

func TestHierarchyStoreCreateAndList(t *testing.T) {
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	conn, err := Open(testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := Ping(ctx, conn); err != nil {
		t.Fatalf("failed to ping test DB: %v", err)
	}

	migrator := NewMigrator(conn, "../../../migrations")
	if _, err := migrator.Up(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	store := NewHierarchyStore(conn)
	suffix := time.Now().UnixNano()

	league, err := store.CreateLeague(ctx, hierarchy.CreateLeagueInput{
		Name: fmt.Sprintf("League %d", suffix),
		Slug: fmt.Sprintf("league-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create league: %v", err)
	}

	franchise, err := store.CreateFranchise(ctx, hierarchy.CreateFranchiseInput{
		LeagueID: league.ID,
		Name:     fmt.Sprintf("Franchise %d", suffix),
		Slug:     fmt.Sprintf("franchise-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create franchise: %v", err)
	}

	club, err := store.CreateClub(ctx, hierarchy.CreateClubInput{
		FranchiseID: franchise.ID,
		Name:        fmt.Sprintf("Club %d", suffix),
		Slug:        fmt.Sprintf("club-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create club: %v", err)
	}

	leagues, err := store.ListLeagues(ctx)
	if err != nil {
		t.Fatalf("failed to list leagues: %v", err)
	}
	if len(leagues) == 0 {
		t.Fatal("expected leagues to contain at least one row")
	}

	franchises, err := store.ListFranchises(ctx)
	if err != nil {
		t.Fatalf("failed to list franchises: %v", err)
	}
	if len(franchises) == 0 {
		t.Fatal("expected franchises to contain at least one row")
	}

	clubs, err := store.ListClubs(ctx)
	if err != nil {
		t.Fatalf("failed to list clubs: %v", err)
	}
	if len(clubs) == 0 {
		t.Fatal("expected clubs to contain at least one row")
	}

	team, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{
		ClubID: club.ID,
		Name:   fmt.Sprintf("Team %d", suffix),
		Slug:   fmt.Sprintf("team-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	_, err = store.CreatePlayer(ctx, hierarchy.CreatePlayerInput{
		DisplayName: fmt.Sprintf("Player %d", suffix),
		Slug:        fmt.Sprintf("player-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create player: %v", err)
	}

	playerTwo, err := store.CreatePlayer(ctx, hierarchy.CreatePlayerInput{
		DisplayName: fmt.Sprintf("Player Two %d", suffix),
		Slug:        fmt.Sprintf("player-two-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create second player: %v", err)
	}

	_, err = store.CreateRosterMembership(ctx, hierarchy.CreateRosterMembershipInput{
		PlayerID: playerTwo.ID,
		TeamID:   team.ID,
	})
	if err != nil {
		t.Fatalf("failed to create roster membership: %v", err)
	}

	teams, err := store.ListTeams(ctx)
	if err != nil {
		t.Fatalf("failed to list teams: %v", err)
	}
	if len(teams) == 0 {
		t.Fatal("expected teams to contain at least one row")
	}

	players, err := store.ListPlayers(ctx)
	if err != nil {
		t.Fatalf("failed to list players: %v", err)
	}
	if len(players) == 0 {
		t.Fatal("expected players to contain at least one row")
	}

	memberships, err := store.ListRosterMemberships(ctx)
	if err != nil {
		t.Fatalf("failed to list roster memberships: %v", err)
	}
	if len(memberships) == 0 {
		t.Fatal("expected roster memberships to contain at least one row")
	}
}

func TestHierarchyStoreDependencyViolation(t *testing.T) {
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	conn, err := Open(testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := Ping(ctx, conn); err != nil {
		t.Fatalf("failed to ping test DB: %v", err)
	}

	migrator := NewMigrator(conn, "../../../migrations")
	if _, err := migrator.Up(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	store := NewHierarchyStore(conn)
	_, err = store.CreateFranchise(ctx, hierarchy.CreateFranchiseInput{
		LeagueID: 99999999,
		Name:     "Invalid",
		Slug:     "invalid-franchise-dep-test",
	})
	if err == nil {
		t.Fatal("expected dependency violation for missing league")
	}
	if !errors.Is(err, hierarchy.ErrDependency) {
		t.Fatalf("expected dependency error, got: %v", err)
	}

	_, err = store.CreateRosterMembership(ctx, hierarchy.CreateRosterMembershipInput{
		PlayerID: 99999999,
		TeamID:   99999998,
	})
	if err == nil {
		t.Fatal("expected dependency violation for missing roster membership parents")
	}
	if !errors.Is(err, hierarchy.ErrDependency) {
		t.Fatalf("expected dependency error for roster membership, got: %v", err)
	}
}

func TestHierarchyStoreRosterMembershipConflict(t *testing.T) {
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}

	conn, err := Open(testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to open test DB: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := Ping(ctx, conn); err != nil {
		t.Fatalf("failed to ping test DB: %v", err)
	}

	migrator := NewMigrator(conn, "../../../migrations")
	if _, err := migrator.Up(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	store := NewHierarchyStore(conn)
	suffix := time.Now().UnixNano()

	league, err := store.CreateLeague(ctx, hierarchy.CreateLeagueInput{
		Name: fmt.Sprintf("League Conflict %d", suffix),
		Slug: fmt.Sprintf("league-conflict-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create league: %v", err)
	}

	franchise, err := store.CreateFranchise(ctx, hierarchy.CreateFranchiseInput{
		LeagueID: league.ID,
		Name:     fmt.Sprintf("Franchise Conflict %d", suffix),
		Slug:     fmt.Sprintf("franchise-conflict-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create franchise: %v", err)
	}

	club, err := store.CreateClub(ctx, hierarchy.CreateClubInput{
		FranchiseID: franchise.ID,
		Name:        fmt.Sprintf("Club Conflict %d", suffix),
		Slug:        fmt.Sprintf("club-conflict-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create club: %v", err)
	}

	team, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{
		ClubID: club.ID,
		Name:   fmt.Sprintf("Team Conflict %d", suffix),
		Slug:   fmt.Sprintf("team-conflict-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	player, err := store.CreatePlayer(ctx, hierarchy.CreatePlayerInput{
		DisplayName: fmt.Sprintf("Player Conflict %d", suffix),
		Slug:        fmt.Sprintf("player-conflict-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create player: %v", err)
	}

	_, err = store.CreateRosterMembership(ctx, hierarchy.CreateRosterMembershipInput{
		PlayerID: player.ID,
		TeamID:   team.ID,
	})
	if err != nil {
		t.Fatalf("failed to create initial roster membership: %v", err)
	}

	teamTwo, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{
		ClubID: club.ID,
		Name:   fmt.Sprintf("Team Conflict Two %d", suffix),
		Slug:   fmt.Sprintf("team-conflict-two-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create second team: %v", err)
	}

	_, err = store.CreateRosterMembership(ctx, hierarchy.CreateRosterMembershipInput{
		PlayerID: player.ID,
		TeamID:   teamTwo.ID,
	})
	if err == nil {
		t.Fatal("expected conflict for second active roster membership")
	}
	if !errors.Is(err, hierarchy.ErrConflict) {
		t.Fatalf("expected conflict error for second active roster membership, got: %v", err)
	}
}
