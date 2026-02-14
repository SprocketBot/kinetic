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

	queue, err := store.CreateQueue(ctx, hierarchy.CreateQueueInput{
		Name: fmt.Sprintf("Queue %d", suffix),
		Slug: fmt.Sprintf("queue-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	_, err = store.EnqueueTeam(ctx, hierarchy.EnqueueTeamInput{
		QueueID: queue.ID,
		TeamID:  team.ID,
	})
	if err != nil {
		t.Fatalf("failed to enqueue team: %v", err)
	}

	queues, err := store.ListQueues(ctx)
	if err != nil {
		t.Fatalf("failed to list queues: %v", err)
	}
	if len(queues) == 0 {
		t.Fatal("expected queues to contain at least one row")
	}

	queueEntries, err := store.ListActiveQueueEntries(ctx)
	if err != nil {
		t.Fatalf("failed to list queue entries: %v", err)
	}
	if len(queueEntries) == 0 {
		t.Fatal("expected queue entries to contain at least one row")
	}

	season, err := store.CreateSeason(ctx, hierarchy.CreateSeasonInput{
		Name: fmt.Sprintf("Season %d", suffix),
		Slug: fmt.Sprintf("season-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create season: %v", err)
	}

	group, err := store.CreateScheduleGroup(ctx, hierarchy.CreateScheduleGroupInput{
		SeasonID: season.ID,
		Name:     fmt.Sprintf("Week %d", suffix),
		Sequence: 1,
	})
	if err != nil {
		t.Fatalf("failed to create schedule group: %v", err)
	}

	clubTwo, err := store.CreateClub(ctx, hierarchy.CreateClubInput{
		FranchiseID: franchise.ID,
		Name:        fmt.Sprintf("Club Two %d", suffix),
		Slug:        fmt.Sprintf("club-two-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create second club: %v", err)
	}

	teamTwo, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{
		ClubID: clubTwo.ID,
		Name:   fmt.Sprintf("Team Two %d", suffix),
		Slug:   fmt.Sprintf("team-two-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create second team: %v", err)
	}

	fixture, err := store.CreateFixture(ctx, hierarchy.CreateFixtureInput{
		ScheduleGroupID: group.ID,
		HomeClubID:      club.ID,
		AwayClubID:      clubTwo.ID,
	})
	if err != nil {
		t.Fatalf("failed to create fixture: %v", err)
	}

	_, err = store.CreateMatch(ctx, hierarchy.CreateMatchInput{
		FixtureID:  fixture.ID,
		HomeTeamID: team.ID,
		AwayTeamID: teamTwo.ID,
		State:      "planned",
	})
	if err != nil {
		t.Fatalf("failed to create match: %v", err)
	}

	seasons, err := store.ListSeasons(ctx)
	if err != nil {
		t.Fatalf("failed to list seasons: %v", err)
	}
	if len(seasons) == 0 {
		t.Fatal("expected seasons to contain at least one row")
	}

	groups, err := store.ListScheduleGroups(ctx)
	if err != nil {
		t.Fatalf("failed to list schedule groups: %v", err)
	}
	if len(groups) == 0 {
		t.Fatal("expected schedule groups to contain at least one row")
	}

	fixtures, err := store.ListFixtures(ctx)
	if err != nil {
		t.Fatalf("failed to list fixtures: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("expected fixtures to contain at least one row")
	}

	matches, err := store.ListMatches(ctx)
	if err != nil {
		t.Fatalf("failed to list matches: %v", err)
	}
	if len(matches) == 0 {
		t.Fatal("expected matches to contain at least one row")
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

	_, err = store.EnqueueTeam(ctx, hierarchy.EnqueueTeamInput{
		QueueID: 99999999,
		TeamID:  99999998,
	})
	if err == nil {
		t.Fatal("expected dependency violation for missing queue/team")
	}
	if !errors.Is(err, hierarchy.ErrDependency) {
		t.Fatalf("expected dependency error for queue entry, got: %v", err)
	}

	_, err = store.CreateScheduleGroup(ctx, hierarchy.CreateScheduleGroupInput{
		SeasonID: 99999999,
		Name:     "Invalid Group",
		Sequence: 1,
	})
	if err == nil {
		t.Fatal("expected dependency violation for missing season")
	}
	if !errors.Is(err, hierarchy.ErrDependency) {
		t.Fatalf("expected dependency error for schedule group, got: %v", err)
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

func TestHierarchyStoreQueueEntryConflictAndLeave(t *testing.T) {
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
		Name: fmt.Sprintf("League Queue %d", suffix),
		Slug: fmt.Sprintf("league-queue-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create league: %v", err)
	}
	franchise, err := store.CreateFranchise(ctx, hierarchy.CreateFranchiseInput{
		LeagueID: league.ID,
		Name:     fmt.Sprintf("Franchise Queue %d", suffix),
		Slug:     fmt.Sprintf("franchise-queue-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create franchise: %v", err)
	}
	club, err := store.CreateClub(ctx, hierarchy.CreateClubInput{
		FranchiseID: franchise.ID,
		Name:        fmt.Sprintf("Club Queue %d", suffix),
		Slug:        fmt.Sprintf("club-queue-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create club: %v", err)
	}
	team, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{
		ClubID: club.ID,
		Name:   fmt.Sprintf("Team Queue %d", suffix),
		Slug:   fmt.Sprintf("team-queue-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}

	queue, err := store.CreateQueue(ctx, hierarchy.CreateQueueInput{
		Name: fmt.Sprintf("Queue Queue %d", suffix),
		Slug: fmt.Sprintf("queue-queue-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	_, err = store.EnqueueTeam(ctx, hierarchy.EnqueueTeamInput{
		QueueID: queue.ID,
		TeamID:  team.ID,
	})
	if err != nil {
		t.Fatalf("failed to enqueue team: %v", err)
	}

	_, err = store.EnqueueTeam(ctx, hierarchy.EnqueueTeamInput{
		QueueID: queue.ID,
		TeamID:  team.ID,
	})
	if err == nil {
		t.Fatal("expected conflict for duplicate active queue entry")
	}
	if !errors.Is(err, hierarchy.ErrConflict) {
		t.Fatalf("expected conflict for duplicate queue entry, got: %v", err)
	}

	entry, err := store.LeaveQueue(ctx, hierarchy.LeaveQueueInput{
		QueueID: queue.ID,
		TeamID:  team.ID,
	})
	if err != nil {
		t.Fatalf("failed to leave queue: %v", err)
	}
	if entry.IsActive {
		t.Fatal("expected queue entry to be inactive after leave")
	}
	if entry.LeftAt == nil {
		t.Fatal("expected leftAt to be set after leave")
	}

	_, err = store.LeaveQueue(ctx, hierarchy.LeaveQueueInput{
		QueueID: queue.ID,
		TeamID:  team.ID,
	})
	if err == nil {
		t.Fatal("expected conflict when leaving queue twice")
	}
	if !errors.Is(err, hierarchy.ErrConflict) {
		t.Fatalf("expected conflict when leaving queue twice, got: %v", err)
	}
}

func TestHierarchyStoreMatchReadyValidation(t *testing.T) {
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
		Name: fmt.Sprintf("League Match %d", suffix),
		Slug: fmt.Sprintf("league-match-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create league: %v", err)
	}
	franchise, err := store.CreateFranchise(ctx, hierarchy.CreateFranchiseInput{
		LeagueID: league.ID,
		Name:     fmt.Sprintf("Franchise Match %d", suffix),
		Slug:     fmt.Sprintf("franchise-match-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create franchise: %v", err)
	}
	clubA, err := store.CreateClub(ctx, hierarchy.CreateClubInput{
		FranchiseID: franchise.ID,
		Name:        fmt.Sprintf("Club Match A %d", suffix),
		Slug:        fmt.Sprintf("club-match-a-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create club A: %v", err)
	}
	clubB, err := store.CreateClub(ctx, hierarchy.CreateClubInput{
		FranchiseID: franchise.ID,
		Name:        fmt.Sprintf("Club Match B %d", suffix),
		Slug:        fmt.Sprintf("club-match-b-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create club B: %v", err)
	}
	teamA, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{
		ClubID: clubA.ID,
		Name:   fmt.Sprintf("Team Match A %d", suffix),
		Slug:   fmt.Sprintf("team-match-a-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create team A: %v", err)
	}
	teamB, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{
		ClubID: clubB.ID,
		Name:   fmt.Sprintf("Team Match B %d", suffix),
		Slug:   fmt.Sprintf("team-match-b-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create team B: %v", err)
	}
	season, err := store.CreateSeason(ctx, hierarchy.CreateSeasonInput{
		Name: fmt.Sprintf("Season Match %d", suffix),
		Slug: fmt.Sprintf("season-match-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create season: %v", err)
	}
	group, err := store.CreateScheduleGroup(ctx, hierarchy.CreateScheduleGroupInput{
		SeasonID: season.ID,
		Name:     fmt.Sprintf("Week Match %d", suffix),
		Sequence: 1,
	})
	if err != nil {
		t.Fatalf("failed to create schedule group: %v", err)
	}
	fixture, err := store.CreateFixture(ctx, hierarchy.CreateFixtureInput{
		ScheduleGroupID: group.ID,
		HomeClubID:      clubA.ID,
		AwayClubID:      clubB.ID,
	})
	if err != nil {
		t.Fatalf("failed to create fixture: %v", err)
	}

	_, err = store.CreateMatch(ctx, hierarchy.CreateMatchInput{
		FixtureID:  fixture.ID,
		HomeTeamID: teamA.ID,
		AwayTeamID: teamB.ID,
		State:      "ready",
	})
	if err == nil {
		t.Fatal("expected invalid input for ready match without ratified schedule")
	}
	if !errors.Is(err, hierarchy.ErrInvalidInput) {
		t.Fatalf("expected invalid input for ready match without ratified schedule, got: %v", err)
	}
}
