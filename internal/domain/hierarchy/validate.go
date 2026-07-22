package hierarchy

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
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

func ValidateCreateGameInput(input CreateGameInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("%w: game name is required", ErrInvalidInput)
	}
	if !slugPattern.MatchString(input.Slug) {
		return fmt.Errorf("%w: game slug must be lowercase kebab-case", ErrInvalidInput)
	}
	return nil
}

func ValidateUpsertUserInput(input UpsertUserInput) error {
	if strings.TrimSpace(input.Subject) == "" {
		return fmt.Errorf("%w: user subject is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.DisplayName) == "" {
		return fmt.Errorf("%w: user displayName is required", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateUserPlayerInput(input CreateUserPlayerInput) error {
	if input.UserID <= 0 {
		return fmt.Errorf("%w: userId must be greater than zero", ErrInvalidInput)
	}
	if input.GameID <= 0 {
		return fmt.Errorf("%w: gameId must be greater than zero", ErrInvalidInput)
	}
	return ValidateCreatePlayerInput(CreatePlayerInput{DisplayName: input.DisplayName, Slug: input.Slug})
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

func ValidateAssignRoleInput(input AssignRoleInput) error {
	if input.ActorPlayerID <= 0 {
		return fmt.Errorf("%w: actorPlayerId must be greater than zero", ErrInvalidInput)
	}
	if input.TargetPlayerID <= 0 {
		return fmt.Errorf("%w: targetPlayerId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Reason) == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidInput)
	}

	role := strings.TrimSpace(input.Role)
	switch role {
	case "fm":
		if input.FranchiseID == nil || *input.FranchiseID <= 0 {
			return fmt.Errorf("%w: franchiseId must be provided for fm role", ErrInvalidInput)
		}
		if input.ClubID != nil || input.TeamID != nil {
			return fmt.Errorf("%w: fm role must not include clubId or teamId", ErrInvalidInput)
		}
	case "gm", "agm":
		if input.ClubID == nil || *input.ClubID <= 0 {
			return fmt.Errorf("%w: clubId must be provided for gm/agm role", ErrInvalidInput)
		}
		if input.FranchiseID != nil || input.TeamID != nil {
			return fmt.Errorf("%w: gm/agm role must not include franchiseId or teamId", ErrInvalidInput)
		}
	case "captain":
		if input.TeamID == nil || *input.TeamID <= 0 {
			return fmt.Errorf("%w: teamId must be provided for captain role", ErrInvalidInput)
		}
		if input.FranchiseID != nil || input.ClubID != nil {
			return fmt.Errorf("%w: captain role must not include franchiseId or clubId", ErrInvalidInput)
		}
	default:
		return fmt.Errorf("%w: role must be fm, gm, agm, or captain", ErrInvalidInput)
	}

	return nil
}

func ValidateRevokeRoleInput(input RevokeRoleInput) error {
	if input.ActorPlayerID <= 0 {
		return fmt.Errorf("%w: actorPlayerId must be greater than zero", ErrInvalidInput)
	}
	if input.AssignmentID <= 0 {
		return fmt.Errorf("%w: assignmentId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Reason) == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidInput)
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

func ValidateLinkPlatformAccountInput(input LinkPlatformAccountInput) error {
	if strings.TrimSpace(input.Subject) == "" {
		return fmt.Errorf("%w: subject is required", ErrInvalidInput)
	}
	switch strings.TrimSpace(input.Provider) {
	case "steam", "xbox", "psn", "epic":
	default:
		return fmt.Errorf("%w: provider must be steam, xbox, psn, or epic", ErrInvalidInput)
	}
	if strings.TrimSpace(input.ProviderAccountID) == "" {
		return fmt.Errorf("%w: providerAccountId is required", ErrInvalidInput)
	}
	return nil
}

func ValidateUnlinkPlatformAccountInput(input UnlinkPlatformAccountInput) error {
	if strings.TrimSpace(input.Subject) == "" {
		return fmt.Errorf("%w: subject is required", ErrInvalidInput)
	}
	switch strings.TrimSpace(input.Provider) {
	case "steam", "xbox", "psn", "epic":
	default:
		return fmt.Errorf("%w: provider must be steam, xbox, psn, or epic", ErrInvalidInput)
	}
	if strings.TrimSpace(input.ProviderAccountID) == "" {
		return fmt.Errorf("%w: providerAccountId is required", ErrInvalidInput)
	}
	return nil
}

func ValidateGetEligibilityInput(input GetEligibilityInput) error {
	if strings.TrimSpace(input.Subject) == "" {
		return fmt.Errorf("%w: subject is required", ErrInvalidInput)
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

func ValidateBanPlayerFromQueueInput(input BanPlayerFromQueueInput) error {
	if input.QueueID <= 0 {
		return fmt.Errorf("%w: queueId must be greater than zero", ErrInvalidInput)
	}
	if input.PlayerID <= 0 {
		return fmt.Errorf("%w: playerId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Actor) == "" {
		return fmt.Errorf("%w: actor is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Reason) == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidInput)
	}
	return nil
}

func ValidateUnbanPlayerFromQueueInput(input UnbanPlayerFromQueueInput) error {
	if input.QueueID <= 0 {
		return fmt.Errorf("%w: queueId must be greater than zero", ErrInvalidInput)
	}
	if input.PlayerID <= 0 {
		return fmt.Errorf("%w: playerId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Actor) == "" {
		return fmt.Errorf("%w: actor is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Reason) == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidInput)
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

func ValidateAdjustPlayerRatingInput(input AdjustPlayerRatingInput) error {
	if input.ActorPlayerID <= 0 {
		return fmt.Errorf("%w: actorPlayerId must be greater than zero", ErrInvalidInput)
	}
	if input.TargetPlayerID <= 0 {
		return fmt.Errorf("%w: targetPlayerId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.ContextKey) == "" {
		return fmt.Errorf("%w: contextKey is required", ErrInvalidInput)
	}
	if input.Rating < 0 {
		return fmt.Errorf("%w: rating must be greater than or equal to zero", ErrInvalidInput)
	}
	if input.Uncertainty < 0 {
		return fmt.Errorf("%w: uncertainty must be greater than or equal to zero", ErrInvalidInput)
	}
	if input.MatchesPlayed < 0 {
		return fmt.Errorf("%w: matchesPlayed must be greater than or equal to zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Reason) == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidInput)
	}
	return nil
}

func ValidateCreateResultSubmissionInput(input CreateResultSubmissionInput) error {
	switch strings.TrimSpace(input.ContextType) {
	case "scrim", "match":
	default:
		return fmt.Errorf("%w: contextType must be scrim or match", ErrInvalidInput)
	}
	switch normalizeGameKey(input.GameKey) {
	case "rocket_league":
	default:
		return fmt.Errorf("%w: gameKey must be rocket_league", ErrInvalidInput)
	}
	if input.ContextID <= 0 {
		return fmt.Errorf("%w: contextId must be greater than zero", ErrInvalidInput)
	}
	if input.SubmittedByTeamID <= 0 {
		return fmt.Errorf("%w: submittedByTeamId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.SubmittedBySubject) == "" {
		return fmt.Errorf("%w: submittedBySubject is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.SubmittedByDisplayName) == "" {
		return fmt.Errorf("%w: submittedByDisplayName is required", ErrInvalidInput)
	}
	if input.WinningTeamID <= 0 || input.LosingTeamID <= 0 {
		return fmt.Errorf("%w: winningTeamId and losingTeamId must be greater than zero", ErrInvalidInput)
	}
	if input.WinningTeamID == input.LosingTeamID {
		return fmt.Errorf("%w: winningTeamId and losingTeamId must differ", ErrInvalidInput)
	}
	if err := ValidateRocketLeagueResultPayload(input.PayloadJSON, input.WinningTeamID, input.LosingTeamID); err != nil {
		return err
	}
	return nil
}

func ValidateOverrideResultSubmissionInput(input OverrideResultSubmissionInput) error {
	if input.SubmissionID <= 0 {
		return fmt.Errorf("%w: submissionId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Actor) == "" {
		return fmt.Errorf("%w: actor is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Reason) == "" {
		return fmt.Errorf("%w: reason is required", ErrInvalidInput)
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
	if strings.TrimSpace(input.RatifiedBySubject) == "" {
		return fmt.Errorf("%w: ratifiedBySubject is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.RatifiedByDisplayName) == "" {
		return fmt.Errorf("%w: ratifiedByDisplayName is required", ErrInvalidInput)
	}
	return nil
}

type RocketLeagueScore struct {
	Home int
	Away int
}

func NormalizeGameKey(gameKey string) string {
	return normalizeGameKey(gameKey)
}

func normalizeGameKey(gameKey string) string {
	trimmed := strings.TrimSpace(gameKey)
	if trimmed == "" {
		return "rocket_league"
	}
	return trimmed
}

func ValidateRocketLeagueResultPayload(raw json.RawMessage, winningTeamID, losingTeamID int64) error {
	if len(raw) == 0 {
		return nil
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fmt.Errorf("%w: payloadJson must be valid JSON", ErrInvalidInput)
	}
	score, ok, err := ExtractRocketLeagueScore(raw)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if score.Home == score.Away {
		return fmt.Errorf("%w: rocket league score cannot be tied", ErrInvalidInput)
	}
	if err := validateOptionalSummaryStats(payload["summaryStats"]); err != nil {
		return err
	}
	winnerIsHome := score.Home > score.Away
	if winnerIsHome && winningTeamID == losingTeamID {
		return fmt.Errorf("%w: winningTeamId and losingTeamId must differ", ErrInvalidInput)
	}
	return nil
}

func ExtractRocketLeagueScore(raw json.RawMessage) (RocketLeagueScore, bool, error) {
	if len(raw) == 0 {
		return RocketLeagueScore{}, false, nil
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return RocketLeagueScore{}, false, fmt.Errorf("%w: payloadJson must be valid JSON", ErrInvalidInput)
	}
	if scoreValue, ok := payload["score"]; ok {
		return parseScoreValue(scoreValue)
	}
	if scoreValue, ok := payload["series"]; ok {
		return parseScoreValue(scoreValue)
	}
	home, homeOK, err := intFromAny(payload["homeScore"])
	if err != nil {
		return RocketLeagueScore{}, false, err
	}
	away, awayOK, err := intFromAny(payload["awayScore"])
	if err != nil {
		return RocketLeagueScore{}, false, err
	}
	if homeOK || awayOK {
		if !homeOK || !awayOK {
			return RocketLeagueScore{}, false, fmt.Errorf("%w: homeScore and awayScore must both be present", ErrInvalidInput)
		}
		return RocketLeagueScore{Home: home, Away: away}, true, nil
	}
	return RocketLeagueScore{}, false, nil
}

func parseScoreValue(value any) (RocketLeagueScore, bool, error) {
	switch typed := value.(type) {
	case string:
		parts := strings.Split(typed, "-")
		if len(parts) != 2 {
			return RocketLeagueScore{}, false, fmt.Errorf("%w: score must use home-away format", ErrInvalidInput)
		}
		home, err := strconv.Atoi(strings.TrimSpace(parts[0]))
		if err != nil {
			return RocketLeagueScore{}, false, fmt.Errorf("%w: home score must be an integer", ErrInvalidInput)
		}
		away, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return RocketLeagueScore{}, false, fmt.Errorf("%w: away score must be an integer", ErrInvalidInput)
		}
		return validateScore(home, away)
	case map[string]any:
		home, homeOK, err := intFromAny(typed["home"])
		if err != nil {
			return RocketLeagueScore{}, false, err
		}
		away, awayOK, err := intFromAny(typed["away"])
		if err != nil {
			return RocketLeagueScore{}, false, err
		}
		if !homeOK || !awayOK {
			return RocketLeagueScore{}, false, fmt.Errorf("%w: score.home and score.away are required", ErrInvalidInput)
		}
		return validateScore(home, away)
	default:
		return RocketLeagueScore{}, false, fmt.Errorf("%w: score must be a string or object", ErrInvalidInput)
	}
}

func validateScore(home, away int) (RocketLeagueScore, bool, error) {
	if home < 0 || away < 0 {
		return RocketLeagueScore{}, false, fmt.Errorf("%w: score values must be nonnegative", ErrInvalidInput)
	}
	return RocketLeagueScore{Home: home, Away: away}, true, nil
}

func intFromAny(value any) (int, bool, error) {
	if value == nil {
		return 0, false, nil
	}
	switch typed := value.(type) {
	case float64:
		if typed < 0 || typed != float64(int(typed)) {
			return 0, false, fmt.Errorf("%w: numeric fields must be nonnegative integers", ErrInvalidInput)
		}
		return int(typed), true, nil
	case int:
		if typed < 0 {
			return 0, false, fmt.Errorf("%w: numeric fields must be nonnegative integers", ErrInvalidInput)
		}
		return typed, true, nil
	default:
		return 0, false, fmt.Errorf("%w: numeric fields must be nonnegative integers", ErrInvalidInput)
	}
}

func validateOptionalSummaryStats(value any) error {
	if value == nil {
		return nil
	}
	stats, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("%w: summaryStats must be an object", ErrInvalidInput)
	}
	for key, statValue := range stats {
		if _, _, err := intFromAny(statValue); err != nil {
			return fmt.Errorf("%w: summaryStats.%s must be a nonnegative integer", ErrInvalidInput, key)
		}
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

func ValidateReportExceptionInput(input ReportExceptionInput) error {
	switch strings.TrimSpace(input.Category) {
	case "scheduling_conflict", "no_show", "forfeit", "result_dispute", "replay_parse_failure", "replay_identity_mismatch", "roster_eligibility":
	default:
		return fmt.Errorf("%w: unsupported category", ErrInvalidInput)
	}
	switch strings.TrimSpace(input.ContextType) {
	case "match", "scrim", "result_submission", "replay_evidence":
	default:
		return fmt.Errorf("%w: unsupported contextType", ErrInvalidInput)
	}
	if input.ContextID <= 0 {
		return fmt.Errorf("%w: contextId must be greater than zero", ErrInvalidInput)
	}
	if input.ReportedByTeamID != nil && *input.ReportedByTeamID <= 0 {
		return fmt.Errorf("%w: reportedByTeamId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.ReasonCode) == "" {
		return fmt.Errorf("%w: reasonCode is required", ErrInvalidInput)
	}
	if input.Severity < 1 || input.Severity > 5 {
		return fmt.Errorf("%w: severity must be in range 1-5", ErrInvalidInput)
	}
	if strings.TrimSpace(input.SuggestedAction) == "" {
		return fmt.Errorf("%w: suggestedAction is required", ErrInvalidInput)
	}
	return nil
}

func ValidateTriageExceptionInput(input TriageExceptionInput) error {
	if input.TicketID <= 0 {
		return fmt.Errorf("%w: ticketId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Actor) == "" {
		return fmt.Errorf("%w: actor is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.ReasonCode) == "" {
		return fmt.Errorf("%w: reasonCode is required", ErrInvalidInput)
	}
	if input.Severity < 1 || input.Severity > 5 {
		return fmt.Errorf("%w: severity must be in range 1-5", ErrInvalidInput)
	}
	if strings.TrimSpace(input.SuggestedAction) == "" {
		return fmt.Errorf("%w: suggestedAction is required", ErrInvalidInput)
	}
	if input.MinutesSpent < 0 {
		return fmt.Errorf("%w: minutesSpent must be greater than or equal to zero", ErrInvalidInput)
	}
	return nil
}

func ValidateResolveExceptionInput(input ResolveExceptionInput) error {
	if input.TicketID <= 0 {
		return fmt.Errorf("%w: ticketId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Actor) == "" {
		return fmt.Errorf("%w: actor is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.ResolutionCode) == "" {
		return fmt.Errorf("%w: resolutionCode is required", ErrInvalidInput)
	}
	if input.MinutesSpent < 0 {
		return fmt.Errorf("%w: minutesSpent must be greater than or equal to zero", ErrInvalidInput)
	}
	return nil
}

func ValidateEvaluateSchedulingExceptionInput(input EvaluateSchedulingExceptionInput) error {
	if input.MatchID <= 0 {
		return fmt.Errorf("%w: matchId must be greater than zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.ConflictCode) == "" {
		return fmt.Errorf("%w: conflictCode is required", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Actor) == "" {
		return fmt.Errorf("%w: actor is required", ErrInvalidInput)
	}
	return nil
}

func ValidateEvaluateNoShowExceptionInput(input EvaluateNoShowExceptionInput) error {
	if input.MatchID <= 0 {
		return fmt.Errorf("%w: matchId must be greater than zero", ErrInvalidInput)
	}
	if input.GraceMinutes < 0 {
		return fmt.Errorf("%w: graceMinutes must be greater than or equal to zero", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Actor) == "" {
		return fmt.Errorf("%w: actor is required", ErrInvalidInput)
	}
	return nil
}

func ValidateEvaluateReplayDisputeExceptionInput(input EvaluateReplayDisputeExceptionInput) error {
	if input.ResultSubmissionID <= 0 {
		return fmt.Errorf("%w: resultSubmissionId must be greater than zero", ErrInvalidInput)
	}
	switch strings.TrimSpace(input.ParseStatus) {
	case "parsed", "failed":
	default:
		return fmt.Errorf("%w: parseStatus must be parsed or failed", ErrInvalidInput)
	}
	switch strings.TrimSpace(input.IdentityStatus) {
	case "resolved", "mismatch":
	default:
		return fmt.Errorf("%w: identityStatus must be resolved or mismatch", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Actor) == "" {
		return fmt.Errorf("%w: actor is required", ErrInvalidInput)
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
