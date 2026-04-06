package entity

import "errors"

var (
	ErrEmptyUserID      = errors.New("user_id cannot be empty")
	ErrEmptyKey         = errors.New("setting key cannot be empty")
	ErrSettingNotFound  = errors.New("user setting not found")
)
