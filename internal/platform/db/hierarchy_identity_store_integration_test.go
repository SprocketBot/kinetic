package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kineticbot/kinetic-v3/internal/domain/hierarchy"
)

func TestHierarchyStoreUserGamePlayers(t *testing.T) {
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
	if _, err := NewMigrator(conn, "../../../migrations").Up(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	store := NewHierarchyStore(conn)
	suffix := time.Now().UnixNano()
	user, err := store.UpsertUser(ctx, hierarchy.UpsertUserInput{
		Subject:     fmt.Sprintf("local:jake-%d", suffix),
		DisplayName: "Jake",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	updated, err := store.UpsertUser(ctx, hierarchy.UpsertUserInput{Subject: user.Subject, DisplayName: "Jake Updated"})
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}
	if updated.ID != user.ID || updated.DisplayName != "Jake Updated" {
		t.Fatalf("expected same user with refreshed display name, got %#v", updated)
	}

	trackmania, err := store.CreateGame(ctx, hierarchy.CreateGameInput{
		Name: fmt.Sprintf("Trackmania %d", suffix),
		Slug: fmt.Sprintf("trackmania-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create game: %v", err)
	}
	games, err := store.ListGames(ctx)
	if err != nil || len(games) == 0 {
		t.Fatalf("expected games including seeded Rocket League: games=%#v err=%v", games, err)
	}

	rocketLeaguePlayer, err := store.CreateUserPlayer(ctx, hierarchy.CreateUserPlayerInput{
		UserID: user.ID, GameID: games[0].ID,
		DisplayName: "Rocket League Jake", Slug: fmt.Sprintf("rocket-league-jake-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create Rocket League player: %v", err)
	}
	if rocketLeaguePlayer.Player.ID == 0 || rocketLeaguePlayer.Game.ID != games[0].ID {
		t.Fatalf("unexpected Rocket League user player: %#v", rocketLeaguePlayer)
	}

	if _, err := store.CreateUserPlayer(ctx, hierarchy.CreateUserPlayerInput{
		UserID: user.ID, GameID: trackmania.ID,
		DisplayName: "Trackmania Jake", Slug: fmt.Sprintf("trackmania-jake-%d", suffix),
	}); err != nil {
		t.Fatalf("failed to create Trackmania player: %v", err)
	}
	if _, err := store.CreateUserPlayer(ctx, hierarchy.CreateUserPlayerInput{
		UserID: user.ID, GameID: games[0].ID,
		DisplayName: "Second Rocket League Jake", Slug: fmt.Sprintf("second-rocket-league-jake-%d", suffix),
	}); !errors.Is(err, hierarchy.ErrConflict) {
		t.Fatalf("expected one-player-per-game conflict, got %v", err)
	}

	players, err := store.ListUserPlayers(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to list user players: %v", err)
	}
	if len(players) != 2 {
		t.Fatalf("expected two game-specific players, got %#v", players)
	}
}
