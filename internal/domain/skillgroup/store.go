package skillgroup

import "context"

type Store interface {
	CreateSkillGroup(ctx context.Context, input CreateSkillGroupInput) (SkillGroup, error)
	ListSkillGroups(ctx context.Context, leagueID int64) ([]SkillGroup, error)
	GetSkillGroup(ctx context.Context, skillGroupID int64) (SkillGroup, error)
	UpdateSkillGroup(ctx context.Context, input UpdateSkillGroupInput) (SkillGroup, error)
	// GetSkillGroupForRating returns the active skill group whose floor <= rating <= ceiling for a league.
	// If multiple groups overlap at this rating, returns the one with the highest display_order.
	// If rating exceeds all ceilings, returns the group with the highest rating_ceiling.
	GetSkillGroupForRating(ctx context.Context, leagueID int64, rating int32) (SkillGroup, error)
	// RecordTransition inserts a skill_group_transition row and returns it.
	RecordTransition(ctx context.Context, input RecordTransitionInput) (SkillGroupTransition, error)
	// ListTransitionsByPlayerID returns all skill group transitions for a player, most recent first.
	ListTransitionsByPlayerID(ctx context.Context, playerID int64) ([]SkillGroupTransition, error)
}
