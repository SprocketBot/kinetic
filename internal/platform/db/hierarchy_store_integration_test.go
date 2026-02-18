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

	if _, err := store.AdvanceQueueEntryStage(ctx, hierarchy.AdvanceQueueEntryStageInput{
		QueueID: queue.ID,
		TeamID:  team.ID,
		Stage:   2,
	}); err != nil {
		t.Fatalf("failed to advance queue stage: %v", err)
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

	if _, err := store.EnqueueTeam(ctx, hierarchy.EnqueueTeamInput{
		QueueID: queue.ID,
		TeamID:  teamTwo.ID,
	}); err != nil {
		t.Fatalf("failed to enqueue second team: %v", err)
	}

	if _, err := store.PromoteQueueToScrim(ctx, hierarchy.PromoteQueueToScrimInput{
		QueueID: queue.ID,
	}); err != nil {
		t.Fatalf("failed to promote queue to scrim: %v", err)
	}

	scrims, err := store.ListScrims(ctx)
	if err != nil {
		t.Fatalf("failed to list scrims: %v", err)
	}
	if len(scrims) == 0 {
		t.Fatal("expected scrims to contain at least one row")
	}

	decisions, err := store.ListMatchmakingDecisions(ctx)
	if err != nil {
		t.Fatalf("failed to list matchmaking decisions: %v", err)
	}
	if len(decisions) == 0 {
		t.Fatal("expected matchmaking decisions to contain at least one row")
	}

	if _, err := conn.ExecContext(
		ctx,
		`INSERT INTO player_ratings(player_id, context_key, rating, uncertainty, matches_played) VALUES ($1, $2, $3, $4, $5)`,
		playerTwo.ID,
		"scrim-3v3",
		1025,
		320,
		7,
	); err != nil {
		t.Fatalf("failed to insert player rating baseline row: %v", err)
	}

	ratings, err := store.ListPlayerRatings(ctx)
	if err != nil {
		t.Fatalf("failed to list player ratings: %v", err)
	}
	if len(ratings) == 0 {
		t.Fatal("expected player ratings to contain at least one row")
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

	_, err = store.CreateScrim(ctx, hierarchy.CreateScrimInput{
		QueueID:    99999999,
		HomeTeamID: 99999998,
		AwayTeamID: 99999997,
		State:      "created",
	})
	if err == nil {
		t.Fatal("expected dependency violation for missing queue/teams on scrim create")
	}
	if !errors.Is(err, hierarchy.ErrDependency) {
		t.Fatalf("expected dependency error for scrim create, got: %v", err)
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

	if _, err := store.AdvanceQueueEntryStage(ctx, hierarchy.AdvanceQueueEntryStageInput{
		QueueID: queue.ID,
		TeamID:  team.ID,
		Stage:   2,
	}); err != nil {
		t.Fatalf("failed to advance queue stage: %v", err)
	}

	_, err = store.AdvanceQueueEntryStage(ctx, hierarchy.AdvanceQueueEntryStageInput{
		QueueID: queue.ID,
		TeamID:  team.ID,
		Stage:   1,
	})
	if err == nil {
		t.Fatal("expected conflict when decreasing queue entry stage")
	}
	if !errors.Is(err, hierarchy.ErrConflict) {
		t.Fatalf("expected conflict when decreasing stage, got: %v", err)
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

	_, err = store.AdvanceQueueEntryStage(ctx, hierarchy.AdvanceQueueEntryStageInput{
		QueueID: queue.ID,
		TeamID:  team.ID,
		Stage:   1,
	})
	if err == nil {
		t.Fatal("expected conflict when advancing stage on inactive queue entry")
	}
	if !errors.Is(err, hierarchy.ErrConflict) {
		t.Fatalf("expected conflict when advancing stage on inactive entry, got: %v", err)
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

func TestHierarchyStorePromoteQueueToScrimRatingFirstOrdering(t *testing.T) {
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
		Name: fmt.Sprintf("League W9 %d", suffix),
		Slug: fmt.Sprintf("league-w9-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create league: %v", err)
	}
	franchise, err := store.CreateFranchise(ctx, hierarchy.CreateFranchiseInput{
		LeagueID: league.ID,
		Name:     fmt.Sprintf("Franchise W9 %d", suffix),
		Slug:     fmt.Sprintf("franchise-w9-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create franchise: %v", err)
	}

	clubA, err := store.CreateClub(ctx, hierarchy.CreateClubInput{
		FranchiseID: franchise.ID,
		Name:        fmt.Sprintf("Club W9 A %d", suffix),
		Slug:        fmt.Sprintf("club-w9-a-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create club A: %v", err)
	}
	clubB, err := store.CreateClub(ctx, hierarchy.CreateClubInput{
		FranchiseID: franchise.ID,
		Name:        fmt.Sprintf("Club W9 B %d", suffix),
		Slug:        fmt.Sprintf("club-w9-b-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create club B: %v", err)
	}
	clubC, err := store.CreateClub(ctx, hierarchy.CreateClubInput{
		FranchiseID: franchise.ID,
		Name:        fmt.Sprintf("Club W9 C %d", suffix),
		Slug:        fmt.Sprintf("club-w9-c-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create club C: %v", err)
	}

	teamA, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{
		ClubID: clubA.ID,
		Name:   fmt.Sprintf("Team W9 A %d", suffix),
		Slug:   fmt.Sprintf("team-w9-a-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create team A: %v", err)
	}
	teamB, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{
		ClubID: clubB.ID,
		Name:   fmt.Sprintf("Team W9 B %d", suffix),
		Slug:   fmt.Sprintf("team-w9-b-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create team B: %v", err)
	}
	teamC, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{
		ClubID: clubC.ID,
		Name:   fmt.Sprintf("Team W9 C %d", suffix),
		Slug:   fmt.Sprintf("team-w9-c-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create team C: %v", err)
	}

	playerA, err := store.CreatePlayer(ctx, hierarchy.CreatePlayerInput{
		DisplayName: fmt.Sprintf("Player W9 A %d", suffix),
		Slug:        fmt.Sprintf("player-w9-a-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create player A: %v", err)
	}
	playerB, err := store.CreatePlayer(ctx, hierarchy.CreatePlayerInput{
		DisplayName: fmt.Sprintf("Player W9 B %d", suffix),
		Slug:        fmt.Sprintf("player-w9-b-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create player B: %v", err)
	}
	playerC, err := store.CreatePlayer(ctx, hierarchy.CreatePlayerInput{
		DisplayName: fmt.Sprintf("Player W9 C %d", suffix),
		Slug:        fmt.Sprintf("player-w9-c-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create player C: %v", err)
	}

	if _, err := store.CreateRosterMembership(ctx, hierarchy.CreateRosterMembershipInput{PlayerID: playerA.ID, TeamID: teamA.ID}); err != nil {
		t.Fatalf("failed to create roster membership A: %v", err)
	}
	if _, err := store.CreateRosterMembership(ctx, hierarchy.CreateRosterMembershipInput{PlayerID: playerB.ID, TeamID: teamB.ID}); err != nil {
		t.Fatalf("failed to create roster membership B: %v", err)
	}
	if _, err := store.CreateRosterMembership(ctx, hierarchy.CreateRosterMembershipInput{PlayerID: playerC.ID, TeamID: teamC.ID}); err != nil {
		t.Fatalf("failed to create roster membership C: %v", err)
	}

	queue, err := store.CreateQueue(ctx, hierarchy.CreateQueueInput{
		Name: fmt.Sprintf("Queue W9 %d", suffix),
		Slug: fmt.Sprintf("queue-w9-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	if _, err := conn.ExecContext(ctx, `INSERT INTO player_ratings(player_id, context_key, rating, uncertainty, matches_played) VALUES ($1, $2, $3, $4, $5)`, playerA.ID, queue.Slug, 1000, 300, 10); err != nil {
		t.Fatalf("failed to seed rating A: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `INSERT INTO player_ratings(player_id, context_key, rating, uncertainty, matches_played) VALUES ($1, $2, $3, $4, $5)`, playerB.ID, queue.Slug, 1010, 300, 10); err != nil {
		t.Fatalf("failed to seed rating B: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `INSERT INTO player_ratings(player_id, context_key, rating, uncertainty, matches_played) VALUES ($1, $2, $3, $4, $5)`, playerC.ID, queue.Slug, 1400, 300, 10); err != nil {
		t.Fatalf("failed to seed rating C: %v", err)
	}

	if _, err := store.EnqueueTeam(ctx, hierarchy.EnqueueTeamInput{QueueID: queue.ID, TeamID: teamA.ID}); err != nil {
		t.Fatalf("failed to enqueue team A: %v", err)
	}
	if _, err := store.EnqueueTeam(ctx, hierarchy.EnqueueTeamInput{QueueID: queue.ID, TeamID: teamB.ID}); err != nil {
		t.Fatalf("failed to enqueue team B: %v", err)
	}
	if _, err := store.EnqueueTeam(ctx, hierarchy.EnqueueTeamInput{QueueID: queue.ID, TeamID: teamC.ID}); err != nil {
		t.Fatalf("failed to enqueue team C: %v", err)
	}

	scrim, err := store.PromoteQueueToScrim(ctx, hierarchy.PromoteQueueToScrimInput{QueueID: queue.ID})
	if err != nil {
		t.Fatalf("failed to promote queue to scrim: %v", err)
	}

	if !((scrim.HomeTeamID == teamA.ID && scrim.AwayTeamID == teamB.ID) || (scrim.HomeTeamID == teamB.ID && scrim.AwayTeamID == teamA.ID)) {
		t.Fatalf("expected A/B pairing from rating-first ordering, got home=%d away=%d", scrim.HomeTeamID, scrim.AwayTeamID)
	}

	decisions, err := store.ListMatchmakingDecisions(ctx)
	if err != nil {
		t.Fatalf("failed to list matchmaking decisions: %v", err)
	}
	if len(decisions) == 0 {
		t.Fatal("expected at least one matchmaking decision")
	}
	decision := decisions[len(decisions)-1]
	if decision.ScrimID != scrim.ID {
		t.Fatalf("expected decision scrim_id %d, got %d", scrim.ID, decision.ScrimID)
	}
	if decision.RatingSpread != 10 {
		t.Fatalf("expected rating spread 10, got %d", decision.RatingSpread)
	}
	if decision.OrderingStrategy == "" {
		t.Fatal("expected non-empty ordering strategy")
	}
}

func TestHierarchyStoreScrimStateTransitions(t *testing.T) {
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

	league, err := store.CreateLeague(ctx, hierarchy.CreateLeagueInput{Name: fmt.Sprintf("League W9 State %d", suffix), Slug: fmt.Sprintf("league-w9-state-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create league: %v", err)
	}
	franchise, err := store.CreateFranchise(ctx, hierarchy.CreateFranchiseInput{LeagueID: league.ID, Name: fmt.Sprintf("Franchise W9 State %d", suffix), Slug: fmt.Sprintf("franchise-w9-state-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create franchise: %v", err)
	}
	clubA, err := store.CreateClub(ctx, hierarchy.CreateClubInput{FranchiseID: franchise.ID, Name: fmt.Sprintf("Club W9 State A %d", suffix), Slug: fmt.Sprintf("club-w9-state-a-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create club A: %v", err)
	}
	clubB, err := store.CreateClub(ctx, hierarchy.CreateClubInput{FranchiseID: franchise.ID, Name: fmt.Sprintf("Club W9 State B %d", suffix), Slug: fmt.Sprintf("club-w9-state-b-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create club B: %v", err)
	}
	teamA, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{ClubID: clubA.ID, Name: fmt.Sprintf("Team W9 State A %d", suffix), Slug: fmt.Sprintf("team-w9-state-a-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create team A: %v", err)
	}
	teamB, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{ClubID: clubB.ID, Name: fmt.Sprintf("Team W9 State B %d", suffix), Slug: fmt.Sprintf("team-w9-state-b-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create team B: %v", err)
	}
	queue, err := store.CreateQueue(ctx, hierarchy.CreateQueueInput{Name: fmt.Sprintf("Queue W9 State %d", suffix), Slug: fmt.Sprintf("queue-w9-state-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	scrim, err := store.CreateScrim(ctx, hierarchy.CreateScrimInput{
		QueueID:    queue.ID,
		HomeTeamID: teamA.ID,
		AwayTeamID: teamB.ID,
		State:      "created",
	})
	if err != nil {
		t.Fatalf("failed to create scrim: %v", err)
	}

	inProgress, err := store.UpdateScrimState(ctx, hierarchy.UpdateScrimStateInput{
		ScrimID: scrim.ID,
		State:   "in_progress",
	})
	if err != nil {
		t.Fatalf("failed to transition scrim to in_progress: %v", err)
	}
	if inProgress.StartedAt == nil {
		t.Fatal("expected started_at to be set on in_progress transition")
	}

	closed, err := store.UpdateScrimState(ctx, hierarchy.UpdateScrimStateInput{
		ScrimID: scrim.ID,
		State:   "closed",
	})
	if err != nil {
		t.Fatalf("failed to transition scrim to closed: %v", err)
	}
	if closed.EndedAt == nil {
		t.Fatal("expected ended_at to be set on closed transition")
	}

	_, err = store.UpdateScrimState(ctx, hierarchy.UpdateScrimStateInput{
		ScrimID: scrim.ID,
		State:   "in_progress",
	})
	if err == nil {
		t.Fatal("expected conflict for closed -> in_progress transition")
	}
	if !errors.Is(err, hierarchy.ErrConflict) {
		t.Fatalf("expected conflict on invalid transition, got: %v", err)
	}
}

func TestHierarchyStoreProcessQueuePromotionsIdempotent(t *testing.T) {
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

	league, err := store.CreateLeague(ctx, hierarchy.CreateLeagueInput{Name: fmt.Sprintf("League W10 %d", suffix), Slug: fmt.Sprintf("league-w10-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create league: %v", err)
	}
	franchise, err := store.CreateFranchise(ctx, hierarchy.CreateFranchiseInput{LeagueID: league.ID, Name: fmt.Sprintf("Franchise W10 %d", suffix), Slug: fmt.Sprintf("franchise-w10-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create franchise: %v", err)
	}
	clubA, err := store.CreateClub(ctx, hierarchy.CreateClubInput{FranchiseID: franchise.ID, Name: fmt.Sprintf("Club W10 A %d", suffix), Slug: fmt.Sprintf("club-w10-a-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create club A: %v", err)
	}
	clubB, err := store.CreateClub(ctx, hierarchy.CreateClubInput{FranchiseID: franchise.ID, Name: fmt.Sprintf("Club W10 B %d", suffix), Slug: fmt.Sprintf("club-w10-b-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create club B: %v", err)
	}
	teamA, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{ClubID: clubA.ID, Name: fmt.Sprintf("Team W10 A %d", suffix), Slug: fmt.Sprintf("team-w10-a-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create team A: %v", err)
	}
	teamB, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{ClubID: clubB.ID, Name: fmt.Sprintf("Team W10 B %d", suffix), Slug: fmt.Sprintf("team-w10-b-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create team B: %v", err)
	}

	queue, err := store.CreateQueue(ctx, hierarchy.CreateQueueInput{Name: fmt.Sprintf("Queue W10 %d", suffix), Slug: fmt.Sprintf("queue-w10-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	if _, err := store.EnqueueTeam(ctx, hierarchy.EnqueueTeamInput{QueueID: queue.ID, TeamID: teamA.ID}); err != nil {
		t.Fatalf("failed to enqueue team A: %v", err)
	}
	if _, err := store.EnqueueTeam(ctx, hierarchy.EnqueueTeamInput{QueueID: queue.ID, TeamID: teamB.ID}); err != nil {
		t.Fatalf("failed to enqueue team B: %v", err)
	}

	first, err := store.ProcessQueuePromotions(ctx, hierarchy.ProcessQueuePromotionsInput{QueueID: queue.ID})
	if err != nil {
		t.Fatalf("failed to process queue promotions first run: %v", err)
	}
	if first.PromotionsCreated != 1 {
		t.Fatalf("expected one promotion on first run, got %d", first.PromotionsCreated)
	}

	second, err := store.ProcessQueuePromotions(ctx, hierarchy.ProcessQueuePromotionsInput{QueueID: queue.ID})
	if err != nil {
		t.Fatalf("failed to process queue promotions second run: %v", err)
	}
	if second.PromotionsCreated != 0 {
		t.Fatalf("expected zero promotions on second run, got %d", second.PromotionsCreated)
	}
	if second.Conflicts == 0 {
		t.Fatal("expected conflict count > 0 on second run")
	}

	runs, err := store.ListPromotionProcessingRuns(ctx)
	if err != nil {
		t.Fatalf("failed to list promotion processing runs: %v", err)
	}
	if len(runs) == 0 {
		t.Fatal("expected at least one promotion processing run")
	}
}

func TestHierarchyStoreResultSubmissionRatificationFlow(t *testing.T) {
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

	league, err := store.CreateLeague(ctx, hierarchy.CreateLeagueInput{Name: fmt.Sprintf("League W11 %d", suffix), Slug: fmt.Sprintf("league-w11-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create league: %v", err)
	}
	franchise, err := store.CreateFranchise(ctx, hierarchy.CreateFranchiseInput{LeagueID: league.ID, Name: fmt.Sprintf("Franchise W11 %d", suffix), Slug: fmt.Sprintf("franchise-w11-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create franchise: %v", err)
	}
	clubA, err := store.CreateClub(ctx, hierarchy.CreateClubInput{FranchiseID: franchise.ID, Name: fmt.Sprintf("Club W11 A %d", suffix), Slug: fmt.Sprintf("club-w11-a-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create club A: %v", err)
	}
	clubB, err := store.CreateClub(ctx, hierarchy.CreateClubInput{FranchiseID: franchise.ID, Name: fmt.Sprintf("Club W11 B %d", suffix), Slug: fmt.Sprintf("club-w11-b-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create club B: %v", err)
	}
	teamA, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{ClubID: clubA.ID, Name: fmt.Sprintf("Team W11 A %d", suffix), Slug: fmt.Sprintf("team-w11-a-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create team A: %v", err)
	}
	teamB, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{ClubID: clubB.ID, Name: fmt.Sprintf("Team W11 B %d", suffix), Slug: fmt.Sprintf("team-w11-b-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create team B: %v", err)
	}
	queue, err := store.CreateQueue(ctx, hierarchy.CreateQueueInput{Name: fmt.Sprintf("Queue W11 %d", suffix), Slug: fmt.Sprintf("queue-w11-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}

	scrim, err := store.CreateScrim(ctx, hierarchy.CreateScrimInput{
		QueueID:    queue.ID,
		HomeTeamID: teamA.ID,
		AwayTeamID: teamB.ID,
		State:      "created",
	})
	if err != nil {
		t.Fatalf("failed to create scrim: %v", err)
	}

	submission, err := store.CreateResultSubmission(ctx, hierarchy.CreateResultSubmissionInput{
		ContextType:       "scrim",
		ContextID:         scrim.ID,
		SubmittedByTeamID: teamA.ID,
		WinningTeamID:     teamA.ID,
		LosingTeamID:      teamB.ID,
		PayloadJSON:       []byte(`{"score":"3-1"}`),
	})
	if err != nil {
		t.Fatalf("failed to create result submission: %v", err)
	}

	submission, err = store.RatifyResultSubmission(ctx, hierarchy.RatifyResultSubmissionInput{
		SubmissionID: submission.ID,
		TeamID:       teamA.ID,
	})
	if err != nil {
		t.Fatalf("failed to ratify submission by home team: %v", err)
	}
	if submission.State != "pending" {
		t.Fatalf("expected pending after first ratification, got %s", submission.State)
	}

	submission, err = store.RatifyResultSubmission(ctx, hierarchy.RatifyResultSubmissionInput{
		SubmissionID: submission.ID,
		TeamID:       teamB.ID,
	})
	if err != nil {
		t.Fatalf("failed to ratify submission by away team: %v", err)
	}
	if submission.State != "ratified" {
		t.Fatalf("expected ratified after both teams ratify, got %s", submission.State)
	}

	_, err = store.RatifyResultSubmission(ctx, hierarchy.RatifyResultSubmissionInput{
		SubmissionID: submission.ID,
		TeamID:       teamB.ID,
	})
	if err == nil {
		t.Fatal("expected conflict when ratifying an already ratified submission")
	}
	if !errors.Is(err, hierarchy.ErrConflict) {
		t.Fatalf("expected conflict on duplicate ratification, got: %v", err)
	}

	submissions, err := store.ListResultSubmissions(ctx)
	if err != nil {
		t.Fatalf("failed to list result submissions: %v", err)
	}
	if len(submissions) == 0 {
		t.Fatal("expected at least one result submission")
	}
}

func TestHierarchyStoreResultOverrideAuditFlow(t *testing.T) {
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

	league, err := store.CreateLeague(ctx, hierarchy.CreateLeagueInput{Name: fmt.Sprintf("League W11 Override %d", suffix), Slug: fmt.Sprintf("league-w11-override-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create league: %v", err)
	}
	franchise, err := store.CreateFranchise(ctx, hierarchy.CreateFranchiseInput{LeagueID: league.ID, Name: fmt.Sprintf("Franchise W11 Override %d", suffix), Slug: fmt.Sprintf("franchise-w11-override-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create franchise: %v", err)
	}
	clubA, err := store.CreateClub(ctx, hierarchy.CreateClubInput{FranchiseID: franchise.ID, Name: fmt.Sprintf("Club W11 Override A %d", suffix), Slug: fmt.Sprintf("club-w11-override-a-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create club A: %v", err)
	}
	clubB, err := store.CreateClub(ctx, hierarchy.CreateClubInput{FranchiseID: franchise.ID, Name: fmt.Sprintf("Club W11 Override B %d", suffix), Slug: fmt.Sprintf("club-w11-override-b-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create club B: %v", err)
	}
	teamA, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{ClubID: clubA.ID, Name: fmt.Sprintf("Team W11 Override A %d", suffix), Slug: fmt.Sprintf("team-w11-override-a-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create team A: %v", err)
	}
	teamB, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{ClubID: clubB.ID, Name: fmt.Sprintf("Team W11 Override B %d", suffix), Slug: fmt.Sprintf("team-w11-override-b-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create team B: %v", err)
	}

	queue, err := store.CreateQueue(ctx, hierarchy.CreateQueueInput{Name: fmt.Sprintf("Queue W11 Override %d", suffix), Slug: fmt.Sprintf("queue-w11-override-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	scrim, err := store.CreateScrim(ctx, hierarchy.CreateScrimInput{
		QueueID:    queue.ID,
		HomeTeamID: teamA.ID,
		AwayTeamID: teamB.ID,
		State:      "created",
	})
	if err != nil {
		t.Fatalf("failed to create scrim: %v", err)
	}

	submission, err := store.CreateResultSubmission(ctx, hierarchy.CreateResultSubmissionInput{
		ContextType:       "scrim",
		ContextID:         scrim.ID,
		SubmittedByTeamID: teamA.ID,
		WinningTeamID:     teamA.ID,
		LosingTeamID:      teamB.ID,
		PayloadJSON:       []byte(`{"score":"3-1"}`),
	})
	if err != nil {
		t.Fatalf("failed to create submission: %v", err)
	}

	overridden, err := store.OverrideResultSubmission(ctx, hierarchy.OverrideResultSubmissionInput{
		SubmissionID:  submission.ID,
		Actor:         "league-admin",
		Reason:        "manual correction",
		WinningTeamID: teamB.ID,
		LosingTeamID:  teamA.ID,
	})
	if err != nil {
		t.Fatalf("failed to override submission: %v", err)
	}
	if overridden.WinningTeamID != teamB.ID {
		t.Fatalf("expected winning team %d after override, got %d", teamB.ID, overridden.WinningTeamID)
	}
	if overridden.State != "ratified" {
		t.Fatalf("expected submission state ratified after override, got %s", overridden.State)
	}

	overrides, err := store.ListResultOverrides(ctx)
	if err != nil {
		t.Fatalf("failed to list result overrides: %v", err)
	}
	if len(overrides) == 0 {
		t.Fatal("expected at least one result override audit row")
	}
	latest := overrides[0]
	if latest.SubmissionID != submission.ID {
		t.Fatalf("expected override submission_id %d, got %d", submission.ID, latest.SubmissionID)
	}
	if latest.PreviousWinningTeamID != teamA.ID || latest.NewWinningTeamID != teamB.ID {
		t.Fatalf("expected audit winner transition %d->%d, got %d->%d", teamA.ID, teamB.ID, latest.PreviousWinningTeamID, latest.NewWinningTeamID)
	}
}

func TestHierarchyStoreReplayIngestionDeduplicatesAndLinksSubmission(t *testing.T) {
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

	league, err := store.CreateLeague(ctx, hierarchy.CreateLeagueInput{Name: fmt.Sprintf("League W12 %d", suffix), Slug: fmt.Sprintf("league-w12-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create league: %v", err)
	}
	franchise, err := store.CreateFranchise(ctx, hierarchy.CreateFranchiseInput{LeagueID: league.ID, Name: fmt.Sprintf("Franchise W12 %d", suffix), Slug: fmt.Sprintf("franchise-w12-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create franchise: %v", err)
	}
	clubA, err := store.CreateClub(ctx, hierarchy.CreateClubInput{FranchiseID: franchise.ID, Name: fmt.Sprintf("Club W12 A %d", suffix), Slug: fmt.Sprintf("club-w12-a-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create club A: %v", err)
	}
	clubB, err := store.CreateClub(ctx, hierarchy.CreateClubInput{FranchiseID: franchise.ID, Name: fmt.Sprintf("Club W12 B %d", suffix), Slug: fmt.Sprintf("club-w12-b-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create club B: %v", err)
	}
	teamA, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{ClubID: clubA.ID, Name: fmt.Sprintf("Team W12 A %d", suffix), Slug: fmt.Sprintf("team-w12-a-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create team A: %v", err)
	}
	teamB, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{ClubID: clubB.ID, Name: fmt.Sprintf("Team W12 B %d", suffix), Slug: fmt.Sprintf("team-w12-b-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create team B: %v", err)
	}
	queue, err := store.CreateQueue(ctx, hierarchy.CreateQueueInput{Name: fmt.Sprintf("Queue W12 %d", suffix), Slug: fmt.Sprintf("queue-w12-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create queue: %v", err)
	}
	scrim, err := store.CreateScrim(ctx, hierarchy.CreateScrimInput{
		QueueID:    queue.ID,
		HomeTeamID: teamA.ID,
		AwayTeamID: teamB.ID,
		State:      "created",
	})
	if err != nil {
		t.Fatalf("failed to create scrim: %v", err)
	}
	submission, err := store.CreateResultSubmission(ctx, hierarchy.CreateResultSubmissionInput{
		ContextType:       "scrim",
		ContextID:         scrim.ID,
		SubmittedByTeamID: teamA.ID,
		WinningTeamID:     teamA.ID,
		LosingTeamID:      teamB.ID,
		PayloadJSON:       []byte(`{"score":"3-1"}`),
	})
	if err != nil {
		t.Fatalf("failed to create result submission: %v", err)
	}

	firstIngest, err := store.IngestReplayEvidence(ctx, hierarchy.IngestReplayEvidenceInput{
		ContextType:        "scrim",
		ContextID:          scrim.ID,
		SubmittedByTeamID:  teamA.ID,
		ReplayBody:         "week12-replay-body",
		ParserName:         "sprocket-rl-parser",
		ParserVersion:      "v0.1.0",
		ParserConfigDigest: "cfg-week12",
		ParseOutputJSON:    []byte(`{"goals":4}`),
		ResultSubmissionID: &submission.ID,
	})
	if err != nil {
		t.Fatalf("failed first replay ingest: %v", err)
	}
	if firstIngest.Duplicate {
		t.Fatal("expected first ingest to be non-duplicate")
	}

	secondIngest, err := store.IngestReplayEvidence(ctx, hierarchy.IngestReplayEvidenceInput{
		ContextType:        "scrim",
		ContextID:          scrim.ID,
		SubmittedByTeamID:  teamA.ID,
		ReplayBody:         "week12-replay-body",
		ParserName:         "sprocket-rl-parser",
		ParserVersion:      "v0.1.0",
		ParserConfigDigest: "cfg-week12",
		ParseOutputJSON:    []byte(`{"goals":4}`),
		ResultSubmissionID: &submission.ID,
	})
	if err != nil {
		t.Fatalf("failed second replay ingest: %v", err)
	}
	if !secondIngest.Duplicate {
		t.Fatal("expected second ingest to be duplicate")
	}
	if secondIngest.Evidence.ID != firstIngest.Evidence.ID {
		t.Fatalf("expected duplicate ingest to reuse evidence ID %d, got %d", firstIngest.Evidence.ID, secondIngest.Evidence.ID)
	}

	evidence, err := store.ListReplayEvidence(ctx)
	if err != nil {
		t.Fatalf("failed to list replay evidence: %v", err)
	}
	if len(evidence) == 0 {
		t.Fatal("expected replay evidence rows")
	}

	parseRuns, err := store.ListReplayParseRuns(ctx)
	if err != nil {
		t.Fatalf("failed to list replay parse runs: %v", err)
	}
	if len(parseRuns) < 2 {
		t.Fatalf("expected at least two parse runs, got %d", len(parseRuns))
	}

	links, err := store.ListResultSubmissionReplayLinks(ctx)
	if err != nil {
		t.Fatalf("failed to list result submission replay links: %v", err)
	}
	if len(links) == 0 {
		t.Fatal("expected replay links rows")
	}
}

func TestHierarchyStoreExceptionAutomationFlow(t *testing.T) {
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

	league, err := store.CreateLeague(ctx, hierarchy.CreateLeagueInput{Name: fmt.Sprintf("League P %d", suffix), Slug: fmt.Sprintf("league-p-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create league: %v", err)
	}
	franchise, err := store.CreateFranchise(ctx, hierarchy.CreateFranchiseInput{LeagueID: league.ID, Name: fmt.Sprintf("Franchise P %d", suffix), Slug: fmt.Sprintf("franchise-p-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create franchise: %v", err)
	}
	clubA, err := store.CreateClub(ctx, hierarchy.CreateClubInput{FranchiseID: franchise.ID, Name: fmt.Sprintf("Club P A %d", suffix), Slug: fmt.Sprintf("club-p-a-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create club A: %v", err)
	}
	clubB, err := store.CreateClub(ctx, hierarchy.CreateClubInput{FranchiseID: franchise.ID, Name: fmt.Sprintf("Club P B %d", suffix), Slug: fmt.Sprintf("club-p-b-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create club B: %v", err)
	}
	teamA, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{ClubID: clubA.ID, Name: fmt.Sprintf("Team P A %d", suffix), Slug: fmt.Sprintf("team-p-a-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create team A: %v", err)
	}
	teamB, err := store.CreateTeam(ctx, hierarchy.CreateTeamInput{ClubID: clubB.ID, Name: fmt.Sprintf("Team P B %d", suffix), Slug: fmt.Sprintf("team-p-b-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create team B: %v", err)
	}
	season, err := store.CreateSeason(ctx, hierarchy.CreateSeasonInput{Name: fmt.Sprintf("Season P %d", suffix), Slug: fmt.Sprintf("season-p-%d", suffix)})
	if err != nil {
		t.Fatalf("failed to create season: %v", err)
	}
	group, err := store.CreateScheduleGroup(ctx, hierarchy.CreateScheduleGroupInput{SeasonID: season.ID, Name: "Week 1", Sequence: 1})
	if err != nil {
		t.Fatalf("failed to create schedule group: %v", err)
	}
	fixture, err := store.CreateFixture(ctx, hierarchy.CreateFixtureInput{ScheduleGroupID: group.ID, HomeClubID: clubA.ID, AwayClubID: clubB.ID})
	if err != nil {
		t.Fatalf("failed to create fixture: %v", err)
	}
	match, err := store.CreateMatch(ctx, hierarchy.CreateMatchInput{
		FixtureID:  fixture.ID,
		HomeTeamID: teamA.ID,
		AwayTeamID: teamB.ID,
		State:      "planned",
	})
	if err != nil {
		t.Fatalf("failed to create match: %v", err)
	}

	reportedByTeamID := teamA.ID
	ticket, err := store.ReportException(ctx, hierarchy.ReportExceptionInput{
		Category:         "scheduling_conflict",
		ContextType:      "match",
		ContextID:        match.ID,
		ReportedByTeamID: &reportedByTeamID,
		ReasonCode:       "time_unavailable",
		Severity:         3,
		SuggestedAction:  "propose_reschedule",
		DetailsJSON:      []byte(`{"source":"integration"}`),
	})
	if err != nil {
		t.Fatalf("failed to report exception: %v", err)
	}

	if _, err := store.TriageException(ctx, hierarchy.TriageExceptionInput{
		TicketID:        ticket.ID,
		Actor:           "ops-user",
		ReasonCode:      "captain_conflict",
		Severity:        2,
		SuggestedAction: "offer_slots",
		MinutesSpent:    5,
	}); err != nil {
		t.Fatalf("failed to triage exception: %v", err)
	}

	if _, err := store.ResolveException(ctx, hierarchy.ResolveExceptionInput{
		TicketID:       ticket.ID,
		Actor:          "ops-user",
		ResolutionCode: "rescheduled",
		Notes:          "captains agreed",
		MinutesSpent:   10,
	}); err != nil {
		t.Fatalf("failed to resolve exception: %v", err)
	}

	scheduling, err := store.EvaluateSchedulingException(ctx, hierarchy.EvaluateSchedulingExceptionInput{
		MatchID:       match.ID,
		ConflictCode:  "captain_conflict",
		HomeConfirmed: false,
		AwayConfirmed: false,
		Actor:         "ops-bot",
	})
	if err != nil {
		t.Fatalf("failed scheduling evaluation: %v", err)
	}
	if scheduling.AutoResolved {
		t.Fatal("expected scheduling evaluation to remain open")
	}

	noShow, err := store.EvaluateNoShowException(ctx, hierarchy.EvaluateNoShowExceptionInput{
		MatchID:       match.ID,
		HomeCheckedIn: true,
		AwayCheckedIn: false,
		GraceMinutes:  20,
		Actor:         "ops-bot",
	})
	if err != nil {
		t.Fatalf("failed no-show evaluation: %v", err)
	}
	if !noShow.AutoResolved {
		t.Fatal("expected no-show evaluation to auto-resolve")
	}

	metrics, err := store.GetExceptionMetrics(ctx)
	if err != nil {
		t.Fatalf("failed to get exception metrics: %v", err)
	}
	if metrics.ManualTouchesPerFixture < 0 {
		t.Fatal("expected non-negative manual touches per fixture")
	}

	actions, err := store.ListExceptionActions(ctx)
	if err != nil {
		t.Fatalf("failed to list exception actions: %v", err)
	}
	if len(actions) == 0 {
		t.Fatal("expected exception actions to be present")
	}

	inbox, err := store.ListOperatorInbox(ctx)
	if err != nil {
		t.Fatalf("failed to list operator inbox: %v", err)
	}
	if len(inbox) == 0 {
		t.Fatal("expected operator inbox to contain open/triaged tickets")
	}
}

func TestHierarchyStoreAdjustPlayerRatingAuditAndGuardrail(t *testing.T) {
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

	actor, err := store.CreatePlayer(ctx, hierarchy.CreatePlayerInput{
		DisplayName: fmt.Sprintf("Rating Actor %d", suffix),
		Slug:        fmt.Sprintf("rating-actor-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create actor player: %v", err)
	}
	target, err := store.CreatePlayer(ctx, hierarchy.CreatePlayerInput{
		DisplayName: fmt.Sprintf("Rating Target %d", suffix),
		Slug:        fmt.Sprintf("rating-target-%d", suffix),
	})
	if err != nil {
		t.Fatalf("failed to create target player: %v", err)
	}

	updated, err := store.AdjustPlayerRating(ctx, hierarchy.AdjustPlayerRatingInput{
		ActorPlayerID:  actor.ID,
		TargetPlayerID: target.ID,
		ContextKey:     "scrim-3v3",
		Rating:         1115,
		Uncertainty:    210,
		MatchesPlayed:  18,
		Reason:         "manual correction after replay review",
	})
	if err != nil {
		t.Fatalf("failed to adjust player rating: %v", err)
	}
	if updated.PlayerID != target.ID {
		t.Fatalf("expected updated rating player_id %d, got %d", target.ID, updated.PlayerID)
	}
	if updated.Rating != 1115 {
		t.Fatalf("expected updated rating 1115, got %d", updated.Rating)
	}

	adjustments, err := store.ListRatingAdjustments(ctx)
	if err != nil {
		t.Fatalf("failed to list rating adjustments: %v", err)
	}
	if len(adjustments) == 0 {
		t.Fatal("expected at least one rating adjustment audit row")
	}
	latest := adjustments[0]
	if latest.ActorPlayerID != actor.ID {
		t.Fatalf("expected actor_player_id %d, got %d", actor.ID, latest.ActorPlayerID)
	}
	if latest.TargetPlayerID != target.ID {
		t.Fatalf("expected target_player_id %d, got %d", target.ID, latest.TargetPlayerID)
	}
	if latest.NewRating != 1115 {
		t.Fatalf("expected new rating 1115, got %d", latest.NewRating)
	}

	_, err = store.AdjustPlayerRating(ctx, hierarchy.AdjustPlayerRatingInput{
		ActorPlayerID:  actor.ID,
		TargetPlayerID: actor.ID,
		ContextKey:     "scrim-3v3",
		Rating:         1200,
		Uncertainty:    180,
		MatchesPlayed:  19,
		Reason:         "self edit should fail",
	})
	if err == nil {
		t.Fatal("expected self-edit guardrail conflict")
	}
	if !errors.Is(err, hierarchy.ErrConflict) {
		t.Fatalf("expected conflict for self-edit guardrail, got: %v", err)
	}
}
