package domain

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

// SettingUpdated is published when a site setting value changes.
// Subscribers should invalidate caches.
type SettingUpdated struct {
	domain.BaseEvent
	Key   string
	Value string
}

func NewSettingUpdated(settingID, key, value string) SettingUpdated {
	return SettingUpdated{
		BaseEvent: domain.NewBaseEvent("sitesetting.updated", settingID),
		Key:       key,
		Value:     value,
	}
}

// SettingCreated is published when a new site setting is added.
type SettingCreated struct {
	domain.BaseEvent
	Key string
}

func NewSettingCreated(settingID, key string) SettingCreated {
	return SettingCreated{
		BaseEvent: domain.NewBaseEvent("sitesetting.created", settingID),
		Key:       key,
	}
}
