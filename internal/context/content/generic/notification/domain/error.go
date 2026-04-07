package domain

import "errors"

var (
	ErrEmptyUserID        = errors.New("user_id cannot be empty")
	ErrEmptyTitle         = errors.New("title cannot be empty")
	ErrEmptyMessage       = errors.New("message cannot be empty")
	ErrNotificationNotFound = errors.New("notification not found")
)
