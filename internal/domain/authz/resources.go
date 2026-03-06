package authz

type Resource = string
type Action = string

const (
	ResourceLeague               Resource = "league"
	ResourceFranchise            Resource = "franchise"
	ResourceClub                 Resource = "club"
	ResourceTeam                 Resource = "team"
	ResourcePlayer               Resource = "player"
	ResourceRosterMembership     Resource = "roster_membership"
	ResourceRoleAssignment       Resource = "role_assignment"
	ResourceQueue                Resource = "queue"
	ResourceQueueEntry           Resource = "queue_entry"
	ResourceQueueBan             Resource = "queue_ban"
	ResourceScrim                Resource = "scrim"
	ResourceScrimPromotion       Resource = "scrim_promotion"
	ResourceResultSubmission     Resource = "result_submission"
	ResourceResultOverride       Resource = "result_override"
	ResourceReplayEvidence       Resource = "replay_evidence"
	ResourcePlatformAccountLink  Resource = "platform_account_link"
	ResourcePlayerRating         Resource = "player_rating"
	ResourceRatingAdjustment     Resource = "rating_adjustment"
	ResourceMatchmakingDecision  Resource = "matchmaking_decision"
	ResourceExceptionTicket      Resource = "exception_ticket"
	ResourceSeason               Resource = "season"
	ResourceScheduleGroup        Resource = "schedule_group"
	ResourceFixture              Resource = "fixture"
	ResourceMatch                Resource = "match"
	ResourceEligibility          Resource = "eligibility"
	ResourcePlayerStat           Resource = "player_stat"
	ResourceSkillGroup           Resource = "skill_group"
	ResourceSkillGroupTransition Resource = "skill_group_transition"
	ResourceOrganizationConfig   Resource = "organization_config"
	ResourceAPIToken             Resource = "api_token"
	ResourcePlayerNotification   Resource = "player_notification"
)

const (
	ActionRead     Action = "read"
	ActionCreate   Action = "create"
	ActionUpdate   Action = "update"
	ActionDelete   Action = "delete"
	ActionRatify   Action = "ratify"
	ActionReject   Action = "reject"
	ActionReset    Action = "reset"
	ActionRevoke   Action = "revoke"
	ActionLift     Action = "lift"
	ActionTriage   Action = "triage"
	ActionResolve  Action = "resolve"
	ActionMarkRead Action = "mark_read"
)
