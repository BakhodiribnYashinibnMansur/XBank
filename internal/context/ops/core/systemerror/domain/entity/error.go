package entity

import "errors"

var (
	ErrEmptyCode       = errors.New("error code cannot be empty")
	ErrAlreadyResolved = errors.New("error is already resolved")
	ErrErrorNotFound   = errors.New("system error not found")
)
