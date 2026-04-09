package ports

import "context"

// NotificationSender provides the ability to send notifications from any BC.
type NotificationSender interface {
	Send(ctx context.Context, req SendNotificationRequest) error
}

// SendNotificationRequest is the input for sending a notification.
type SendNotificationRequest struct {
	UserID  string            `json:"user_id"`
	Type    string            `json:"type"`    // "push", "email", "sms", "in_app"
	Title   string            `json:"title"`
	Body    string            `json:"body"`
	Channel string            `json:"channel"` // delivery channel
	Data    map[string]string `json:"data,omitempty"`
}
