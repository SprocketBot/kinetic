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

	_, err = store.CreateClub(ctx, hierarchy.CreateClubInput{
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
}
