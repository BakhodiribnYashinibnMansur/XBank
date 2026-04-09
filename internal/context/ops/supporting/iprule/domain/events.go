package domain

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

// IPRuleCreated is emitted when a new IP rule is created.
type IPRuleCreated struct {
	domain.BaseEvent
	IPAddress string
	RuleType  string
}

func NewIPRuleCreated(id, ipAddress string, ruleType RuleType) IPRuleCreated {
	return IPRuleCreated{
		BaseEvent: domain.NewBaseEvent("iprule.created", id),
		IPAddress: ipAddress,
		RuleType:  string(ruleType),
	}
}

// IPRuleDeleted is emitted when an IP rule is removed.
type IPRuleDeleted struct {
	domain.BaseEvent
}

func NewIPRuleDeleted(id string) IPRuleDeleted {
	return IPRuleDeleted{
		BaseEvent: domain.NewBaseEvent("iprule.deleted", id),
	}
}
