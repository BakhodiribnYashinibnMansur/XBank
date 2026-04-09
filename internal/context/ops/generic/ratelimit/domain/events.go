package domain

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

// RateLimitRuleCreated is emitted when a new rate limit rule is created.
type RateLimitRuleCreated struct {
	domain.BaseEvent
	Key         string
	MaxRequests int
}

func NewRateLimitRuleCreated(id, key string, maxRequests int) RateLimitRuleCreated {
	return RateLimitRuleCreated{
		BaseEvent:   domain.NewBaseEvent("ratelimit.rule.created", id),
		Key:         key,
		MaxRequests: maxRequests,
	}
}

// RateLimitRuleDeleted is emitted when a rate limit rule is removed.
type RateLimitRuleDeleted struct {
	domain.BaseEvent
}

func NewRateLimitRuleDeleted(id string) RateLimitRuleDeleted {
	return RateLimitRuleDeleted{
		BaseEvent: domain.NewBaseEvent("ratelimit.rule.deleted", id),
	}
}
