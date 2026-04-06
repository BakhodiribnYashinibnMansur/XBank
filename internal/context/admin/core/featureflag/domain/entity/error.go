package entity

import "errors"

var (
	ErrEmptyKey           = errors.New("feature flag key cannot be empty")
	ErrFlagNotFound       = errors.New("feature flag not found")
	ErrKeyExists          = errors.New("feature flag key already exists")
	ErrInvalidRolloutPct  = errors.New("rollout percentage must be 0-100")
	ErrEmptyRuleGroupName = errors.New("rule group name cannot be empty")
	ErrRuleGroupNotFound  = errors.New("rule group not found")
)
