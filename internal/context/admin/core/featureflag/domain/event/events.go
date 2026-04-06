package event

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

// FlagCreated is published when a new feature flag is created.
type FlagCreated struct {
	domain.BaseEvent
	Key string
}

func NewFlagCreated(flagID, key string) FlagCreated {
	return FlagCreated{
		BaseEvent: domain.NewBaseEvent("featureflag.created", flagID),
		Key:       key,
	}
}

// FlagUpdated is published when a feature flag is modified.
// Subscribers should invalidate cached evaluations.
type FlagUpdated struct {
	domain.BaseEvent
	Key string
}

func NewFlagUpdated(flagID, key string) FlagUpdated {
	return FlagUpdated{
		BaseEvent: domain.NewBaseEvent("featureflag.updated", flagID),
		Key:       key,
	}
}

// FlagDeleted is published when a feature flag is removed.
type FlagDeleted struct {
	domain.BaseEvent
	Key string
}

func NewFlagDeleted(flagID, key string) FlagDeleted {
	return FlagDeleted{
		BaseEvent: domain.NewBaseEvent("featureflag.deleted", flagID),
		Key:       key,
	}
}
