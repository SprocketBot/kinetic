package hierarchy

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

func ValidateCreateLeagueInput(input CreateLeagueInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: league name is required", ErrInvalidInput)
	}
	if !slugPattern.MatchString(input.Slug) {
		return fmt.Errorf("%w: league slug must be lowercase kebab-case", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateFranchiseInput(input CreateFranchiseInput) error {
	if input.LeagueID <= 0 {
		return fmt.Errorf("%w: leagueId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: franchise name is required", ErrInvalidInput)
	}
	if !slugPattern.MatchString(input.Slug) {
		return fmt.Errorf("%w: franchise slug must be lowercase kebab-case", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateClubInput(input CreateClubInput) error {
	if input.FranchiseID <= 0 {
		return fmt.Errorf("%w: franchiseId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: club name is required", ErrInvalidInput)
	}
	if !slugPattern.MatchString(input.Slug) {
		return fmt.Errorf("%w: club slug must be lowercase kebab-case", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateTeamInput(input CreateTeamInput) error {
	if input.ClubID <= 0 {
		return fmt.Errorf("%w: clubId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: team name is required", ErrInvalidInput)
	}
	if !slugPattern.MatchString(input.Slug) {
		return fmt.Errorf("%w: team slug must be lowercase kebab-case", ErrInvalidInput)
	}
	return nil
}

func ValidateCreatePlayerInput(input CreatePlayerInput) error {
	if strings.TrimSpace(input.DisplayName) == "" {
		return fmt.Errorf("%w: player displayName is required", ErrInvalidInput)
	}
	if !slugPattern.MatchString(input.Slug) {
		return fmt.Errorf("%w: player slug must be lowercase kebab-case", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateRosterMembershipInput(input CreateRosterMembershipInput) error {
	if input.PlayerID <= 0 {
		return fmt.Errorf("%w: playerId must be greater than zero", ErrInvalidInput)
	}
	if input.TeamID <= 0 {
		return fmt.Errorf("%w: teamId must be greater than zero", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateQueueInput(input CreateQueueInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: queue name is required", ErrInvalidInput)
	}
	if !slugPattern.MatchString(input.Slug) {
		return fmt.Errorf("%w: queue slug must be lowercase kebab-case", ErrInvalidInput)
	}
	return nil
}

func ValidateEnqueueTeamInput(input EnqueueTeamInput) error {
	if input.QueueID <= 0 {
		return fmt.Errorf("%w: queueId must be greater than zero", ErrInvalidInput)
	}
	if input.TeamID <= 0 {
		return fmt.Errorf("%w: teamId must be greater than zero", ErrInvalidInput)
	}
	return nil
}

func ValidateLeaveQueueInput(input LeaveQueueInput) error {
	if input.QueueID <= 0 {
		return fmt.Errorf("%w: queueId must be greater than zero", ErrInvalidInput)
	}
	if input.TeamID <= 0 {
		return fmt.Errorf("%w: teamId must be greater than zero", ErrInvalidInput)
	}
	return nil
}

func ValidateAdvanceQueueEntryStageInput(input AdvanceQueueEntryStageInput) error {
	if input.QueueID <= 0 {
		return fmt.Errorf("%w: queueId must be greater than zero", ErrInvalidInput)
	}
	if input.TeamID <= 0 {
		return fmt.Errorf("%w: teamId must be greater than zero", ErrInvalidInput)
	}
	if input.Stage < 1 {
		return fmt.Errorf("%w: stage must be greater than or equal to one", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateScrimInput(input CreateScrimInput) error {
	if input.QueueID <= 0 {
		return fmt.Errorf("%w: queueId must be greater than zero", ErrInvalidInput)
	}
	if input.HomeTeamID <= 0 {
		return fmt.Errorf("%w: homeTeamId must be greater than zero", ErrInvalidInput)
	}
	if input.AwayTeamID <= 0 {
		return fmt.Errorf("%w: awayTeamId must be greater than zero", ErrInvalidInput)
	}
	if input.HomeTeamID == input.AwayTeamID {
		return fmt.Errorf("%w: scrim homeTeamId and awayTeamId must differ", ErrInvalidInput)
	}

	state := strings.TrimSpace(input.State)
	switch state {
	case "created", "in_progress", "closed", "voided":
		return nil
	default:
		return fmt.Errorf("%w: scrim state must be created, in_progress, closed, or voided", ErrInvalidInput)
	}
}

func ValidatePromoteQueueToScrimInput(input PromoteQueueToScrimInput) error {
	if input.QueueID <= 0 {
		return fmt.Errorf("%w: queueId must be greater than zero", ErrInvalidInput)
	}
	return nil
}

func ValidateUpdateScrimStateInput(input UpdateScrimStateInput) error {
	if input.ScrimID <= 0 {
		return fmt.Errorf("%w: scrimId must be greater than zero", ErrInvalidInput)
	}

	state := strings.TrimSpace(input.State)
	switch state {
	case "in_progress", "closed", "voided":
		return nil
	default:
		return fmt.Errorf("%w: scrim state transition target must be in_progress, closed, or voided", ErrInvalidInput)
	}
}

func ValidateProcessQueuePromotionsInput(input ProcessQueuePromotionsInput) error {
	if input.QueueID < 0 {
		return fmt.Errorf("%w: queueId must be zero (all queues) or greater", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateResultSubmissionInput(input CreateResultSubmissionInput) error {
	switch strings.TrimSpace(input.ContextType) {
	case "scrim", "match":
	default:
		return fmt.Errorf("%w: contextType must be scrim or match", ErrInvalidInput)
	}
	if input.ContextID <= 0 {
		return fmt.Errorf("%w: contextId must be greater than zero", ErrInvalidInput)
	}
	if input.SubmittedByTeamID <= 0 {
		return fmt.Errorf("%w: submittedByTeamId must be greater than zero", ErrInvalidInput)
	}
	if input.WinningTeamID <= 0 || input.LosingTeamID <= 0 {
		return fmt.Errorf("%w: winningTeamId and losingTeamId must be greater than zero", ErrInvalidInput)
	}
	if input.WinningTeamID == input.LosingTeamID {
		return fmt.Errorf("%w: winningTeamId and losingTeamId must differ", ErrInvalidInput)
	}
	return nil
}

func ValidateRatifyResultSubmissionInput(input RatifyResultSubmissionInput) error {
	if input.SubmissionID <= 0 {
		return fmt.Errorf("%w: submissionId must be greater than zero", ErrInvalidInput)
	}
	if input.TeamID <= 0 {
		return fmt.Errorf("%w: teamId must be greater than zero", ErrInvalidInput)
	}
	return nil
}

func ValidateRejectResultSubmissionInput(input RejectResultSubmissionInput) error {
	if input.SubmissionID <= 0 {
		return fmt.Errorf("%w: submissionId must be greater than zero", ErrInvalidInput)
	}
	if input.TeamID <= 0 {
		return fmt.Errorf("%w: teamId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Reason) == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidInput)
	}
	return nil
}

func ValidateIngestReplayEvidenceInput(input IngestReplayEvidenceInput) error {
	switch strings.TrimSpace(input.ContextType) {
	case "scrim", "match":
	default:
		return fmt.Errorf("%w: contextType must be scrim or match", ErrInvalidInput)
	}
	if input.ContextID <= 0 {
		return fmt.Errorf("%w: contextId must be greater than zero", ErrInvalidInput)
	}
	if input.SubmittedByTeamID <= 0 {
		return fmt.Errorf("%w: submittedByTeamId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.ReplayBody) == "" {
		return fmt.Errorf("%w: replayBody is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.ParserName) == "" {
		return fmt.Errorf("%w: parserName is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.ParserVersion) == "" {
		return fmt.Errorf("%w: parserVersion is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.ParserConfigDigest) == "" {
		return fmt.Errorf("%w: parserConfigDigest is required", ErrInvalidInput)
	}
	if input.ResultSubmissionID != nil && *input.ResultSubmissionID <= 0 {
		return fmt.Errorf("%w: resultSubmissionId must be greater than zero when provided", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateSeasonInput(input CreateSeasonInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: season name is required", ErrInvalidInput)
	}
	if !slugPattern.MatchString(input.Slug) {
		return fmt.Errorf("%w: season slug must be lowercase kebab-case", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateScheduleGroupInput(input CreateScheduleGroupInput) error {
	if input.SeasonID <= 0 {
		return fmt.Errorf("%w: seasonId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: scheduleGroup name is required", ErrInvalidInput)
	}
	if input.Sequence <= 0 {
		return fmt.Errorf("%w: scheduleGroup sequence must be greater than zero", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateFixtureInput(input CreateFixtureInput) error {
	if input.ScheduleGroupID <= 0 {
		return fmt.Errorf("%w: scheduleGroupId must be greater than zero", ErrInvalidInput)
	}
	if input.HomeClubID <= 0 {
		return fmt.Errorf("%w: homeClubId must be greater than zero", ErrInvalidInput)
	}
	if input.AwayClubID <= 0 {
		return fmt.Errorf("%w: awayClubId must be greater than zero", ErrInvalidInput)
	}
	if input.HomeClubID == input.AwayClubID {
		return fmt.Errorf("%w: fixture homeClubId and awayClubId must differ", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateMatchInput(input CreateMatchInput) error {
	if input.FixtureID <= 0 {
		return fmt.Errorf("%w: fixtureId must be greater than zero", ErrInvalidInput)
	}
	if input.HomeTeamID <= 0 {
		return fmt.Errorf("%w: homeTeamId must be greater than zero", ErrInvalidInput)
	}
	if input.AwayTeamID <= 0 {
		return fmt.Errorf("%w: awayTeamId must be greater than zero", ErrInvalidInput)
	}
	if input.HomeTeamID == input.AwayTeamID {
		return fmt.Errorf("%w: match homeTeamId and awayTeamId must differ", ErrInvalidInput)
	}

	state := strings.TrimSpace(input.State)
	if state == "" {
		return fmt.Errorf("%w: match state is required", ErrInvalidInput)
	}
	switch state {
	case "planned":
		return nil
	case "ready":
		if input.ScheduledFor == nil || input.HomeTimeRatifiedAt == nil || input.AwayTimeRatifiedAt == nil {
			return fmt.Errorf("%w: match ready requires scheduledFor and both ratification timestamps", ErrInvalidInput)
		}
		if !isTimelineValid(*input.ScheduledFor, *input.HomeTimeRatifiedAt, *input.AwayTimeRatifiedAt) {
			return fmt.Errorf("%w: match ready ratification timestamps must be <= scheduledFor", ErrInvalidInput)
		}
		return nil
	default:
		return fmt.Errorf("%w: match state must be planned or ready", ErrInvalidInput)
	}
}

func isTimelineValid(scheduledFor, homeRatifiedAt, awayRatifiedAt time.Time) bool {
	return !homeRatifiedAt.After(scheduledFor) && !awayRatifiedAt.After(scheduledFor)
}
