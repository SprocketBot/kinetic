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
