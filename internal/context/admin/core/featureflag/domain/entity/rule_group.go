package entity

import "time"

// RuleGroup is a set of conditions that must ALL match (AND logic).
// Multiple rule groups on a flag use OR logic between them.
// Priority determines evaluation order (lower = first).
type RuleGroup struct {
	ID         string
	FlagID     string
	Name       string
	Priority   int
	Value      string // value to return if this rule group matches
	Conditions []Condition
	CreatedAt  time.Time
}

// NewRuleGroup creates a rule group with validation.
func NewRuleGroup(flagID, name, value string, priority int) (*RuleGroup, error) {
	if name == "" {
		return nil, ErrEmptyRuleGroupName
	}
	return &RuleGroup{
		FlagID:    flagID,
		Name:      name,
		Priority:  priority,
		Value:     value,
		CreatedAt: time.Now(),
	}, nil
}

// AddCondition adds a condition to the rule group.
func (rg *RuleGroup) AddCondition(c Condition) {
	rg.Conditions = append(rg.Conditions, c)
}

// Matches checks if all conditions in this group match the given attributes.
func (rg *RuleGroup) Matches(attributes map[string]string) bool {
	if len(rg.Conditions) == 0 {
		return false
	}
	for _, c := range rg.Conditions {
		if !c.Matches(attributes) {
			return false
		}
	}
	return true
}
