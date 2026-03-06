package skillgroup

import "time"

type SkillGroupTransition struct {
	ID                 int64     `json:"id"`
	PlayerID           int64     `json:"playerId"`
	FromSkillGroupID   *int64    `json:"fromSkillGroupId,omitempty"`
	ToSkillGroupID     int64     `json:"toSkillGroupId"`
	RatingAtTransition int32     `json:"ratingAtTransition"`
	Direction          string    `json:"direction"`
	EffectiveAt        time.Time `json:"effectiveAt"`
}

type SkillGroup struct {
	ID                 int64     `json:"id"`
	LeagueID           int64     `json:"leagueId"`
	Name               string    `json:"name"`
	DisplayOrder       int32     `json:"displayOrder"`
	RatingFloor        int32     `json:"ratingFloor"`
	RatingCeiling      int32     `json:"ratingCeiling"`
	PromotionThreshold *int32    `json:"promotionThreshold,omitempty"`
	DemotionThreshold  *int32    `json:"demotionThreshold,omitempty"`
	IsActive           bool      `json:"isActive"`
	CreatedAt          time.Time `json:"createdAt"`
}

type CreateSkillGroupInput struct {
	LeagueID           int64  `json:"leagueId"`
	Name               string `json:"name"`
	DisplayOrder       int32  `json:"displayOrder"`
	RatingFloor        int32  `json:"ratingFloor"`
	RatingCeiling      int32  `json:"ratingCeiling"`
	PromotionThreshold *int32 `json:"promotionThreshold,omitempty"`
	DemotionThreshold  *int32 `json:"demotionThreshold,omitempty"`
}

type UpdateSkillGroupInput struct {
	SkillGroupID       int64  `json:"skillGroupId"`
	RatingFloor        *int32 `json:"ratingFloor,omitempty"`
	RatingCeiling      *int32 `json:"ratingCeiling,omitempty"`
	PromotionThreshold *int32 `json:"promotionThreshold,omitempty"`
	DemotionThreshold  *int32 `json:"demotionThreshold,omitempty"`
	IsActive           *bool  `json:"isActive,omitempty"`
}

type RecordTransitionInput struct {
	PlayerID           int64  `json:"playerId"`
	FromSkillGroupID   *int64 `json:"fromSkillGroupId,omitempty"`
	ToSkillGroupID     int64  `json:"toSkillGroupId"`
	RatingAtTransition int32  `json:"ratingAtTransition"`
	Direction          string `json:"direction"`
}
