package entity

import "errors"

var (
	ErrEmptyKey       = errors.New("setting key cannot be empty")
	ErrEmptyValue     = errors.New("setting value cannot be empty")
	ErrSettingNotFound = errors.New("site setting not found")
	ErrKeyExists      = errors.New("setting key already exists")
)
