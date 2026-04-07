package domain

import "errors"

var (
	ErrEmptyUserID       = errors.New("user_id cannot be empty")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrExportNotFound    = errors.New("data export not found")
)
