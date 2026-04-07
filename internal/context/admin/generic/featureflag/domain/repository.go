package domain

import (
	"context"

)

// WriteRepository defines write operations for FeatureFlag aggregate.
type WriteRepository interface {
	Save(ctx context.Context, flag *FeatureFlag) error
	Update(ctx context.Context, flag *FeatureFlag) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*FeatureFlag, error)
	FindByKey(ctx context.Context, key string) (*FeatureFlag, error)
	FindAll(ctx context.Context) ([]*FeatureFlag, error)
}

// RuleGroupRepository manages rule groups and conditions.
type RuleGroupRepository interface {
	SaveRuleGroup(ctx context.Context, rg *RuleGroup) error
	UpdateRuleGroup(ctx context.Context, rg *RuleGroup) error
	DeleteRuleGroup(ctx context.Context, id string) error
	FindRuleGroupsByFlagID(ctx context.Context, flagID string) ([]RuleGroup, error)
}

// Evaluator evaluates feature flags for a given context.
type Evaluator interface {
	// IsEnabled checks if a boolean flag is enabled for the user.
	IsEnabled(ctx context.Context, key string, userID string, attributes map[string]string) bool
	// GetValue returns the resolved string value of a flag.
	GetValue(ctx context.Context, key string, userID string, attributes map[string]string) (string, error)
}

// FeatureFlagView is the read projection.
type FeatureFlagView struct {
	ID           string              `json:"id"`
	Key          string              `json:"key"`
	Description  string              `json:"description"`
	FlagType     FlagType     `json:"flag_type"`
	DefaultValue string              `json:"default_value"`
	Active       bool                `json:"active"`
	RolloutPct   int                 `json:"rollout_pct"`
	RuleGroups   []RuleGroupView     `json:"rule_groups,omitempty"`
	CreatedAt    string              `json:"created_at"`
	UpdatedAt    string              `json:"updated_at"`
}

// RuleGroupView is the read projection for rule groups.
type RuleGroupView struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Priority   int             `json:"priority"`
	Value      string          `json:"value"`
	Conditions []ConditionView `json:"conditions"`
}

// ConditionView is the read projection for conditions.
type ConditionView struct {
	ID        string           `json:"id"`
	Attribute string           `json:"attribute"`
	Operator  Operator  `json:"operator"`
	Value     string           `json:"value"`
}

// FeatureFlagFilter for list queries.
type FeatureFlagFilter struct {
	Key    string
	Active *bool
	Limit  int
	Offset int
}

// ReadRepository defines read operations for FeatureFlag projections.
type ReadRepository interface {
	FindByID(ctx context.Context, id string) (*FeatureFlagView, error)
	List(ctx context.Context, filter FeatureFlagFilter) ([]*FeatureFlagView, int64, error)
}
