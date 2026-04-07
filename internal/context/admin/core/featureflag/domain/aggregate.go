package domain

import (
	"hash/fnv"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// FlagType defines the data type of a feature flag's value.
type FlagType string

const (
	FlagTypeBool   FlagType = "bool"
	FlagTypeString FlagType = "string"
	FlagTypeInt    FlagType = "int"
	FlagTypeFloat  FlagType = "float"
)

// FeatureFlag is the aggregate root for feature toggles.
// Supports rollout percentage, rule-based evaluation, and typed values.
type FeatureFlag struct {
	domain.AggregateRoot
	Key          string
	Description  string
	FlagType     FlagType
	DefaultValue string // string representation of default
	Active       bool
	RolloutPct   int // 0-100, percentage of users who see the feature
	RuleGroups   []RuleGroup
}

// NewFeatureFlag creates a new feature flag.
func NewFeatureFlag(key, description string, flagType FlagType, defaultValue string) (*FeatureFlag, error) {
	if key == "" {
		return nil, ErrEmptyKey
	}
	if flagType == "" {
		flagType = FlagTypeBool
	}

	now := time.Now()
	f := &FeatureFlag{
		Key:          key,
		Description:  description,
		FlagType:     flagType,
		DefaultValue: defaultValue,
		Active:       false,
		RolloutPct:   0,
	}
	f.CreatedAt = now
	f.UpdatedAt = now
	return f, nil
}

// Activate enables the feature flag.
func (f *FeatureFlag) Activate() {
	f.Active = true
	f.Touch()
}

// Deactivate disables the feature flag. Inactive flags always return default.
func (f *FeatureFlag) Deactivate() {
	f.Active = false
	f.Touch()
}

// SetRollout sets the rollout percentage (0-100).
func (f *FeatureFlag) SetRollout(pct int) error {
	if pct < 0 || pct > 100 {
		return ErrInvalidRolloutPct
	}
	f.RolloutPct = pct
	f.Touch()
	return nil
}

// Update modifies the flag's metadata.
func (f *FeatureFlag) Update(description *string, defaultValue *string, active *bool, rolloutPct *int) error {
	if description != nil {
		f.Description = *description
	}
	if defaultValue != nil {
		f.DefaultValue = *defaultValue
	}
	if active != nil {
		f.Active = *active
	}
	if rolloutPct != nil {
		if *rolloutPct < 0 || *rolloutPct > 100 {
			return ErrInvalidRolloutPct
		}
		f.RolloutPct = *rolloutPct
	}
	f.Touch()
	return nil
}

// AddRuleGroup adds a rule group to the flag.
func (f *FeatureFlag) AddRuleGroup(rg RuleGroup) {
	f.RuleGroups = append(f.RuleGroups, rg)
	f.Touch()
}

// IsEnabledForUser checks if the flag is active for a specific user
// based on rollout percentage using deterministic hashing (FNV-1a).
func (f *FeatureFlag) IsEnabledForUser(userID string) bool {
	if !f.Active {
		return false
	}
	if f.RolloutPct >= 100 {
		return true
	}
	if f.RolloutPct <= 0 {
		return false
	}

	// Deterministic bucket: same user always gets same result
	h := fnv.New32a()
	h.Write([]byte(f.Key + ":" + userID))
	bucket := int(h.Sum32() % 100)
	return bucket < f.RolloutPct
}
