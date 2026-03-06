package orgconfig

import "time"

type OrgConfig struct {
	ID        int64     `json:"id"`
	LeagueID  int64     `json:"leagueId"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedBy *int64    `json:"updatedBy,omitempty"`
	UpdatedAt time.Time `json:"updatedAt"`
}

const (
	KeyScrimPopTimeoutMinutes           = "scrim.pop_timeout_minutes"
	KeyScrimQueueBanInitialDurationMins = "scrim.queue_ban_initial_duration_minutes"
	KeyScrimQueueBanDurationModifier    = "scrim.queue_ban_duration_modifier"
	KeyScrimQueueBanModifierFalloffDays = "scrim.queue_ban_modifier_falloff_days"
	KeyScrimRequiredRatifications       = "scrim.required_ratifications"
	KeyQueueExpansionIntervalSeconds    = "queue.expansion_interval_seconds"
)

func Default(key string) string {
	switch key {
	case KeyScrimPopTimeoutMinutes:
		return "5"
	case KeyScrimQueueBanInitialDurationMins:
		return "60"
	case KeyScrimQueueBanDurationModifier:
		return "2.0"
	case KeyScrimQueueBanModifierFalloffDays:
		return "30"
	case KeyScrimRequiredRatifications:
		return "2"
	case KeyQueueExpansionIntervalSeconds:
		return "30"
	default:
		return ""
	}
}

func AllKnownKeys() []string {
	return []string{
		KeyScrimPopTimeoutMinutes,
		KeyScrimQueueBanInitialDurationMins,
		KeyScrimQueueBanDurationModifier,
		KeyScrimQueueBanModifierFalloffDays,
		KeyScrimRequiredRatifications,
		KeyQueueExpansionIntervalSeconds,
	}
}
