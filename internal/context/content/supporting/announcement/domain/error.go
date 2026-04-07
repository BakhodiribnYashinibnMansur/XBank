package domain

import "errors"

var (
	ErrEmptyTitle          = errors.New("at least one language title is required")
	ErrAlreadyPublished    = errors.New("announcement is already published")
	ErrCannotEditPublished = errors.New("cannot edit published announcement")
	ErrAnnouncementNotFound = errors.New("announcement not found")
)
