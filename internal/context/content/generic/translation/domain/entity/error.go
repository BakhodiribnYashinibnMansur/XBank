package entity

import "errors"

var (
	ErrEmptyKey         = errors.New("translation key cannot be empty")
	ErrEmptyValue       = errors.New("translation value cannot be empty")
	ErrEmptyLanguage    = errors.New("language cannot be empty")
	ErrTranslationNotFound = errors.New("translation not found")
	ErrKeyLanguageExists   = errors.New("translation for this key+language already exists")
)
