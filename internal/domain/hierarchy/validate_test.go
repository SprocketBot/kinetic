package hierarchy

import (
	"testing"
	"time"
)

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
		DisplayName: "Player One",
		Slug:        "player-one",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreatePlayerInput(CreatePlayerInput{
		DisplayName: "",
		Slug:        "player-one",
	})
	if err == nil {
		t.Fatal("expected missing displayName to fail validation")
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

func TestValidateCreateQueueInput(t *testing.T) {
	t.Parallel()

	err := ValidateCreateQueueInput(CreateQueueInput{
		Name: "3v3 Ranked",
		Slug: "3v3-ranked",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreateQueueInput(CreateQueueInput{
		Name: "",
		Slug: "3v3-ranked",
	})
	if err == nil {
		t.Fatal("expected missing queue name to fail validation")
	}
}

func TestValidateEnqueueTeamInput(t *testing.T) {
	t.Parallel()

	err := ValidateEnqueueTeamInput(EnqueueTeamInput{
		QueueID: 1,
		TeamID:  2,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateEnqueueTeamInput(EnqueueTeamInput{
		QueueID: 0,
		TeamID:  2,
	})
	if err == nil {
		t.Fatal("expected missing queueId to fail validation")
	}
}

func TestValidateLeaveQueueInput(t *testing.T) {
	t.Parallel()

	err := ValidateLeaveQueueInput(LeaveQueueInput{
		QueueID: 1,
		TeamID:  2,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateLeaveQueueInput(LeaveQueueInput{
		QueueID: 1,
		TeamID:  0,
	})
	if err == nil {
		t.Fatal("expected missing teamId to fail validation")
	}
}

func TestValidateBanPlayerFromQueueInput(t *testing.T) {
	t.Parallel()

	err := ValidateBanPlayerFromQueueInput(BanPlayerFromQueueInput{
		QueueID:  1,
		PlayerID: 2,
		Actor:    "support-operator",
		Reason:   "toxicity",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateBanPlayerFromQueueInput(BanPlayerFromQueueInput{
		QueueID:  1,
		PlayerID: 2,
		Actor:    "support-operator",
		Reason:   "",
	})
	if err == nil {
		t.Fatal("expected missing reason to fail validation")
	}
}

func TestValidateUnbanPlayerFromQueueInput(t *testing.T) {
	t.Parallel()

	err := ValidateUnbanPlayerFromQueueInput(UnbanPlayerFromQueueInput{
		QueueID:  1,
		PlayerID: 2,
		Actor:    "support-operator",
		Reason:   "appeal accepted",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateUnbanPlayerFromQueueInput(UnbanPlayerFromQueueInput{
		QueueID:  0,
		PlayerID: 2,
		Actor:    "support-operator",
		Reason:   "appeal accepted",
	})
	if err == nil {
		t.Fatal("expected missing queueId to fail validation")
	}
}

func TestValidateCreateSeasonInput(t *testing.T) {
	t.Parallel()

	err := ValidateCreateSeasonInput(CreateSeasonInput{
		Name: "Season 1",
		Slug: "season-1",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreateSeasonInput(CreateSeasonInput{
		Name: "Season 1",
		Slug: "Season-1",
	})
	if err == nil {
		t.Fatal("expected invalid season slug to fail validation")
	}
}

func TestValidateCreateScheduleGroupInput(t *testing.T) {
	t.Parallel()

	err := ValidateCreateScheduleGroupInput(CreateScheduleGroupInput{
		SeasonID: 1,
		Name:     "Week 1",
		Sequence: 1,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreateScheduleGroupInput(CreateScheduleGroupInput{
		SeasonID: 1,
		Name:     "Week 1",
		Sequence: 0,
	})
	if err == nil {
		t.Fatal("expected missing sequence to fail validation")
	}
}

func TestValidateCreateFixtureInput(t *testing.T) {
	t.Parallel()

	err := ValidateCreateFixtureInput(CreateFixtureInput{
		ScheduleGroupID: 1,
		HomeClubID:      10,
		AwayClubID:      20,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreateFixtureInput(CreateFixtureInput{
		ScheduleGroupID: 1,
		HomeClubID:      10,
		AwayClubID:      10,
	})
	if err == nil {
		t.Fatal("expected equal clubs to fail validation")
	}
}

func TestValidateCreateMatchInput(t *testing.T) {
	t.Parallel()

	scheduled := time.Now().UTC().Add(24 * time.Hour)
	homeRatified := scheduled.Add(-2 * time.Hour)
	awayRatified := scheduled.Add(-1 * time.Hour)

	err := ValidateCreateMatchInput(CreateMatchInput{
		FixtureID:          1,
		HomeTeamID:         10,
		AwayTeamID:         20,
		State:              "ready",
		ScheduledFor:       &scheduled,
		HomeTimeRatifiedAt: &homeRatified,
		AwayTimeRatifiedAt: &awayRatified,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreateMatchInput(CreateMatchInput{
		FixtureID:  1,
		HomeTeamID: 10,
		AwayTeamID: 20,
		State:      "ready",
	})
	if err == nil {
		t.Fatal("expected ready without scheduling ratification to fail validation")
	}
}

func TestValidateAdvanceQueueEntryStageInput(t *testing.T) {
	t.Parallel()

	err := ValidateAdvanceQueueEntryStageInput(AdvanceQueueEntryStageInput{
		QueueID: 1,
		TeamID:  2,
		Stage:   2,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateAdvanceQueueEntryStageInput(AdvanceQueueEntryStageInput{
		QueueID: 1,
		TeamID:  2,
		Stage:   0,
	})
	if err == nil {
		t.Fatal("expected invalid stage to fail validation")
	}
}

func TestValidateCreateScrimInput(t *testing.T) {
	t.Parallel()

	err := ValidateCreateScrimInput(CreateScrimInput{
		QueueID:    1,
		HomeTeamID: 10,
		AwayTeamID: 20,
		State:      "created",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreateScrimInput(CreateScrimInput{
		QueueID:    1,
		HomeTeamID: 10,
		AwayTeamID: 10,
		State:      "created",
	})
	if err == nil {
		t.Fatal("expected equal teams to fail validation")
	}
}

func TestValidatePromoteQueueToScrimInput(t *testing.T) {
	t.Parallel()

	err := ValidatePromoteQueueToScrimInput(PromoteQueueToScrimInput{
		QueueID: 1,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidatePromoteQueueToScrimInput(PromoteQueueToScrimInput{
		QueueID: 0,
	})
	if err == nil {
		t.Fatal("expected missing queueId to fail validation")
	}
}

func TestValidateUpdateScrimStateInput(t *testing.T) {
	t.Parallel()

	err := ValidateUpdateScrimStateInput(UpdateScrimStateInput{
		ScrimID: 1,
		State:   "in_progress",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateUpdateScrimStateInput(UpdateScrimStateInput{
		ScrimID: 1,
		State:   "created",
	})
	if err == nil {
		t.Fatal("expected invalid target state to fail validation")
	}

	err = ValidateUpdateScrimStateInput(UpdateScrimStateInput{
		ScrimID: 0,
		State:   "closed",
	})
	if err == nil {
		t.Fatal("expected missing scrimId to fail validation")
	}
}

func TestValidateProcessQueuePromotionsInput(t *testing.T) {
	t.Parallel()

	err := ValidateProcessQueuePromotionsInput(ProcessQueuePromotionsInput{
		QueueID: 0,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateProcessQueuePromotionsInput(ProcessQueuePromotionsInput{
		QueueID: -1,
	})
	if err == nil {
		t.Fatal("expected negative queueId to fail validation")
	}
}

func TestValidateAdjustPlayerRatingInput(t *testing.T) {
	t.Parallel()

	err := ValidateAdjustPlayerRatingInput(AdjustPlayerRatingInput{
		ActorPlayerID:  1,
		TargetPlayerID: 2,
		ContextKey:     "scrim-3v3",
		Rating:         1080,
		Uncertainty:    220,
		MatchesPlayed:  32,
		Reason:         "manual correction",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateAdjustPlayerRatingInput(AdjustPlayerRatingInput{
		ActorPlayerID:  0,
		TargetPlayerID: 2,
		ContextKey:     "scrim-3v3",
		Rating:         1080,
		Uncertainty:    220,
		MatchesPlayed:  32,
		Reason:         "manual correction",
	})
	if err == nil {
		t.Fatal("expected missing actorPlayerId to fail validation")
	}
}

func TestValidateCreateResultSubmissionInput(t *testing.T) {
	t.Parallel()

	err := ValidateCreateResultSubmissionInput(CreateResultSubmissionInput{
		ContextType:       "scrim",
		ContextID:         1,
		SubmittedByTeamID: 10,
		WinningTeamID:     10,
		LosingTeamID:      20,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateCreateResultSubmissionInput(CreateResultSubmissionInput{
		ContextType:       "bad",
		ContextID:         1,
		SubmittedByTeamID: 10,
		WinningTeamID:     10,
		LosingTeamID:      20,
	})
	if err == nil {
		t.Fatal("expected invalid contextType to fail validation")
	}
}

func TestValidateOverrideResultSubmissionInput(t *testing.T) {
	t.Parallel()

	err := ValidateOverrideResultSubmissionInput(OverrideResultSubmissionInput{
		SubmissionID:  1,
		Actor:         "league-admin",
		Reason:        "manual correction",
		WinningTeamID: 10,
		LosingTeamID:  20,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateOverrideResultSubmissionInput(OverrideResultSubmissionInput{
		SubmissionID:  1,
		Actor:         "league-admin",
		Reason:        "",
		WinningTeamID: 10,
		LosingTeamID:  20,
	})
	if err == nil {
		t.Fatal("expected missing reason to fail validation")
	}
}

func TestValidateRatifyResultSubmissionInput(t *testing.T) {
	t.Parallel()

	err := ValidateRatifyResultSubmissionInput(RatifyResultSubmissionInput{
		SubmissionID: 1,
		TeamID:       10,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateRatifyResultSubmissionInput(RatifyResultSubmissionInput{
		SubmissionID: 0,
		TeamID:       10,
	})
	if err == nil {
		t.Fatal("expected invalid submissionId to fail validation")
	}
}

func TestValidateRejectResultSubmissionInput(t *testing.T) {
	t.Parallel()

	err := ValidateRejectResultSubmissionInput(RejectResultSubmissionInput{
		SubmissionID: 1,
		TeamID:       10,
		Reason:       "bad replay",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateRejectResultSubmissionInput(RejectResultSubmissionInput{
		SubmissionID: 1,
		TeamID:       10,
		Reason:       "",
	})
	if err == nil {
		t.Fatal("expected missing reason to fail validation")
	}
}

func TestValidateIngestReplayEvidenceInput(t *testing.T) {
	t.Parallel()

	submissionID := int64(1)
	err := ValidateIngestReplayEvidenceInput(IngestReplayEvidenceInput{
		ContextType:        "scrim",
		ContextID:          1,
		SubmittedByTeamID:  10,
		ReplayBody:         "fake-replay-bytes",
		ParserName:         "sprocket-rl-parser",
		ParserVersion:      "v0.1.0",
		ParserConfigDigest: "cfg-001",
		ResultSubmissionID: &submissionID,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateIngestReplayEvidenceInput(IngestReplayEvidenceInput{
		ContextType:        "scrim",
		ContextID:          1,
		SubmittedByTeamID:  10,
		ReplayBody:         "",
		ParserName:         "sprocket-rl-parser",
		ParserVersion:      "v0.1.0",
		ParserConfigDigest: "cfg-001",
	})
	if err == nil {
		t.Fatal("expected empty replayBody to fail validation")
	}
}

func TestValidateReportExceptionInput(t *testing.T) {
	t.Parallel()

	err := ValidateReportExceptionInput(ReportExceptionInput{
		Category:        "scheduling_conflict",
		ContextType:     "match",
		ContextID:       1,
		ReasonCode:      "time_unavailable",
		Severity:        3,
		SuggestedAction: "propose_reschedule",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateReportExceptionInput(ReportExceptionInput{
		Category:        "bad",
		ContextType:     "match",
		ContextID:       1,
		ReasonCode:      "x",
		Severity:        3,
		SuggestedAction: "y",
	})
	if err == nil {
		t.Fatal("expected invalid category to fail validation")
	}
}

func TestValidateTriageExceptionInput(t *testing.T) {
	t.Parallel()

	err := ValidateTriageExceptionInput(TriageExceptionInput{
		TicketID:        1,
		Actor:           "ops-user",
		ReasonCode:      "review_needed",
		Severity:        2,
		SuggestedAction: "contact_captains",
		MinutesSpent:    5,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateTriageExceptionInput(TriageExceptionInput{
		TicketID:        0,
		Actor:           "ops-user",
		ReasonCode:      "review_needed",
		Severity:        2,
		SuggestedAction: "contact_captains",
		MinutesSpent:    5,
	})
	if err == nil {
		t.Fatal("expected invalid ticketId to fail validation")
	}
}

func TestValidateResolveExceptionInput(t *testing.T) {
	t.Parallel()

	err := ValidateResolveExceptionInput(ResolveExceptionInput{
		TicketID:       1,
		Actor:          "ops-user",
		ResolutionCode: "resolved_manual",
		MinutesSpent:   10,
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}

	err = ValidateResolveExceptionInput(ResolveExceptionInput{
		TicketID:       1,
		Actor:          "",
		ResolutionCode: "resolved_manual",
	})
	if err == nil {
		t.Fatal("expected missing actor to fail validation")
	}
}

func TestValidateEvaluateSchedulingExceptionInput(t *testing.T) {
	t.Parallel()

	err := ValidateEvaluateSchedulingExceptionInput(EvaluateSchedulingExceptionInput{
		MatchID:       1,
		ConflictCode:  "captain_conflict",
		HomeConfirmed: false,
		AwayConfirmed: false,
		Actor:         "ops-user",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}
}

func TestValidateEvaluateNoShowExceptionInput(t *testing.T) {
	t.Parallel()

	err := ValidateEvaluateNoShowExceptionInput(EvaluateNoShowExceptionInput{
		MatchID:       1,
		HomeCheckedIn: true,
		AwayCheckedIn: false,
		GraceMinutes:  15,
		Actor:         "ops-user",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}
}

func TestValidateEvaluateReplayDisputeExceptionInput(t *testing.T) {
	t.Parallel()

	err := ValidateEvaluateReplayDisputeExceptionInput(EvaluateReplayDisputeExceptionInput{
		ResultSubmissionID: 1,
		ParseStatus:        "parsed",
		IdentityStatus:     "resolved",
		DisputeRaised:      false,
		Actor:              "ops-user",
	})
	if err != nil {
		t.Fatalf("expected valid input, got error: %v", err)
	}
}
