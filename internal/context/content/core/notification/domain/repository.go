package domain

import (
	"context"

)

type WriteRepository interface {
	Save(ctx context.Context, n *Notification) error
	Update(ctx context.Context, n *Notification) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Notification, error)
}

type NotificationView struct {
	ID        string                   `json:"id"`
	UserID    string                   `json:"user_id"`
	Title     string                   `json:"title"`
	Message   string                   `json:"message"`
	Type      NotificationType  `json:"type"`
	Read      bool                     `json:"read"`
	Data      map[string]string        `json:"data,omitempty"`
	CreatedAt string                   `json:"created_at"`
}

type NotificationFilter struct {
	UserID string
	Type   string
	Unread *bool
	Limit  int
	Offset int
}

type ReadRepository interface {
	FindByID(ctx context.Context, id string) (*NotificationView, error)
	List(ctx context.Context, filter NotificationFilter) ([]*NotificationView, int64, error)
}
