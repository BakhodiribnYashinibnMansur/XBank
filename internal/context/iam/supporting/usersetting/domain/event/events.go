package event

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

type SettingUpdated struct {
	domain.BaseEvent
	UserID string
	Key    string
}

func NewSettingUpdated(id, userID, key string) SettingUpdated {
	return SettingUpdated{BaseEvent: domain.NewBaseEvent("usersetting.updated", id), UserID: userID, Key: key}
}
