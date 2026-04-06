package entity

import "errors"

var (
	ErrEmptyCode        = errors.New("error code cannot be empty")
	ErrCodeNotFound     = errors.New("error code not found")
	ErrCodeAlreadyExists = errors.New("error code already exists")
)
