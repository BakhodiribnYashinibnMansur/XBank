package domain

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

type NotificationSent struct {
	domain.BaseEvent
	UserID string
	Title  string
	Type   string
}

func NewNotificationSent(id, userID, title, notifType string) NotificationSent {
	return NotificationSent{
		BaseEvent: domain.NewBaseEvent("notification.sent", id),
		UserID:    userID,
		Title:     title,
		Type:      notifType,
	}
}
