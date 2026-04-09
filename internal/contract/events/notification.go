package events

import "time"

// NotificationSent is published when a notification is delivered.
type NotificationSent struct {
	NotificationID string    `json:"notification_id"`
	UserID         string    `json:"user_id"`
	Title          string    `json:"title"`
	Type           string    `json:"type"`
	OccurredAt     time.Time `json:"occurred_at"`
}
