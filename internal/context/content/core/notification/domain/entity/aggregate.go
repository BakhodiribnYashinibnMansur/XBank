package entity

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// NotificationType categorizes notifications.
type NotificationType string

const (
	NotificationInfo    NotificationType = "INFO"
	NotificationWarning NotificationType = "WARNING"
	NotificationAlert   NotificationType = "ALERT"
)

// Notification is the aggregate root for user notifications.
// Persisted to DB (unlike SSE-only notifications).
type Notification struct {
	domain.AggregateRoot
	UserID  string
	Title   string
	Message string
	Type    NotificationType
	ReadAt  *time.Time // nil = unread
	Data    map[string]string
}

// NewNotification creates a notification with validation.
func NewNotification(userID, title, message string, notifType NotificationType, data map[string]string) (*Notification, error) {
	if userID == "" {
		return nil, ErrEmptyUserID
	}
	if title == "" {
		return nil, ErrEmptyTitle
	}
	if message == "" {
		return nil, ErrEmptyMessage
	}
	if notifType == "" {
		notifType = NotificationInfo
	}

	now := time.Now()
	n := &Notification{
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notifType,
		Data:    data,
	}
	n.CreatedAt = now
	n.UpdatedAt = now
	return n, nil
}

// MarkAsRead marks the notification as read.
func (n *Notification) MarkAsRead() {
	if n.ReadAt != nil {
		return // already read
	}
	now := time.Now()
	n.ReadAt = &now
	n.Touch()
}

// IsRead returns true if the notification has been read.
func (n *Notification) IsRead() bool {
	return n.ReadAt != nil
}
