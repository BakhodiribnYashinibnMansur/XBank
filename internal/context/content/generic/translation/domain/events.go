package domain

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

// TranslationUpdated is published when a translation value changes.
type TranslationUpdated struct {
	domain.BaseEvent
	Key      string
	Language string
}

func NewTranslationUpdated(id, key, language string) TranslationUpdated {
	return TranslationUpdated{
		BaseEvent: domain.NewBaseEvent("translation.updated", id),
		Key:       key,
		Language:  language,
	}
}
