package entity

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// Language supported by the system.
type Language string

const (
	LangUz Language = "uz"
	LangRu Language = "ru"
	LangEn Language = "en"
)

// Translation is the aggregate root for i18n key-value pairs.
// Natural uniqueness: (key, language). Group allows bulk loading per module.
type Translation struct {
	domain.AggregateRoot
	Key      string
	Language Language
	Value    string
	Group    string // logical grouping: "errors", "auth", "dashboard"
}

// NewTranslation creates a new translation with validation.
func NewTranslation(key string, language Language, value, group string) (*Translation, error) {
	if key == "" {
		return nil, ErrEmptyKey
	}
	if value == "" {
		return nil, ErrEmptyValue
	}
	if language == "" {
		return nil, ErrEmptyLanguage
	}

	now := time.Now()
	t := &Translation{
		Key:      key,
		Language: language,
		Value:    value,
		Group:    group,
	}
	t.CreatedAt = now
	t.UpdatedAt = now
	return t, nil
}

// Update modifies the translation value.
func (t *Translation) Update(value string) error {
	if value == "" {
		return ErrEmptyValue
	}
	t.Value = value
	t.Touch()
	return nil
}
