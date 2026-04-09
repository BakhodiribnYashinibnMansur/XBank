package events

import "time"

// NotificationRequested is published when a notification should be sent.
type NotificationRequested struct {
	NotificationID string            `json:"notification_id"`
	UserID         string            `json:"user_id"`
	Type           string            `json:"type"`
	Title          string            `json:"title"`
	Body           string            `json:"body"`
	Channel        string            `json:"channel"`
	Data           map[string]string `json:"data,omitempty"`
	OccurredAt     time.Time         `json:"occurred_at"`
}

// NotificationSent is published when a notification is delivered.
type NotificationSent struct {
	NotificationID string    `json:"notification_id"`
	UserID         string    `json:"user_id"`
	Title          string    `json:"title"`
	Type           string    `json:"type"`
	OccurredAt     time.Time `json:"occurred_at"`
}
