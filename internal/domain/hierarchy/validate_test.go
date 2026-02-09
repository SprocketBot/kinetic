package hierarchy

import "testing"

func TestValidateCreateLeagueInput(t *testing.T) {
	t.Parallel()

	err := ValidateCreateLeagueInput(CreateLeagueInput{
		Name: "Minor League Esports",
		Slug: "minor-league-esports",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreateLeagueInput(CreateLeagueInput{
		Name: "MLE",
		Slug: "Not-Kebab",
	})
	if err == nil {
		t.Fatal("expected invalid slug to fail validation")
	}
}

func TestValidateCreateFranchiseInput(t *testing.T) {
	t.Parallel()

	err := ValidateCreateFranchiseInput(CreateFranchiseInput{
		LeagueID: 1,
		Name:     "Guardians",
		Slug:     "guardians",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreateFranchiseInput(CreateFranchiseInput{
		LeagueID: 0,
		Name:     "Guardians",
		Slug:     "guardians",
	})
	if err == nil {
		t.Fatal("expected missing leagueId to fail validation")
	}
}

func TestValidateCreateClubInput(t *testing.T) {
	t.Parallel()

	err := ValidateCreateClubInput(CreateClubInput{
		FranchiseID: 1,
		Name:        "Guardians RL",
		Slug:        "guardians-rl",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreateClubInput(CreateClubInput{
		FranchiseID: 1,
		Name:        "",
		Slug:        "guardians-rl",
	})
	if err == nil {
		t.Fatal("expected empty name to fail validation")
	}
}

func TestValidateCreateTeamInput(t *testing.T) {
	t.Parallel()

	err := ValidateCreateTeamInput(CreateTeamInput{
		ClubID: 1,
		Name:   "Team Alpha",
		Slug:   "team-alpha",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreateTeamInput(CreateTeamInput{
		ClubID: 1,
		Name:   "Team Alpha",
		Slug:   "Team-Alpha",
	})
	if err == nil {
		t.Fatal("expected invalid team slug to fail validation")
	}
}

func TestValidateCreatePlayerInput(t *testing.T) {
	t.Parallel()

	err := ValidateCreatePlayerInput(CreatePlayerInput{
		TeamID:      1,
		DisplayName: "Player One",
		Slug:        "player-one",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreatePlayerInput(CreatePlayerInput{
		TeamID:      0,
		DisplayName: "Player One",
		Slug:        "player-one",
	})
	if err == nil {
		t.Fatal("expected missing teamId to fail validation")
	}
}

func TestValidateCreateRosterMembershipInput(t *testing.T) {
	t.Parallel()

	err := ValidateCreateRosterMembershipInput(CreateRosterMembershipInput{
		PlayerID: 1,
		TeamID:   2,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreateRosterMembershipInput(CreateRosterMembershipInput{
		PlayerID: 0,
		TeamID:   2,
	})
	if err == nil {
		t.Fatal("expected missing playerId to fail validation")
	}

	err = ValidateCreateRosterMembershipInput(CreateRosterMembershipInput{
		PlayerID: 1,
		TeamID:   0,
	})
	if err == nil {
		t.Fatal("expected missing teamId to fail validation")
	}
}
