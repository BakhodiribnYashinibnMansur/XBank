package entity

import "errors"

var (
	ErrEmptyName    = errors.New("file name cannot be empty")
	ErrFileNotFound = errors.New("file not found")
)
