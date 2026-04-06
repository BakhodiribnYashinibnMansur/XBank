package entity

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// SettingType categorizes settings for UI grouping.
type SettingType string

const (
	SettingTypeGeneral  SettingType = "general"
	SettingTypeEmail    SettingType = "email"
	SettingTypeSecurity SettingType = "security"
	SettingTypePayment  SettingType = "payment"
)

// SiteSetting is the aggregate root for global key-value configuration.
// Each setting has a unique key and can be categorized by type.
type SiteSetting struct {
	domain.AggregateRoot
	Key         string
	Value       string
	SettingType SettingType
	Description string
}

// NewSiteSetting creates a new site setting with validation.
func NewSiteSetting(key, value string, settingType SettingType, description string) (*SiteSetting, error) {
	if key == "" {
		return nil, ErrEmptyKey
	}
	if value == "" {
		return nil, ErrEmptyValue
	}

	now := time.Now()
	s := &SiteSetting{
		Key:         key,
		Value:       value,
		SettingType: settingType,
		Description: description,
	}
	s.CreatedAt = now
	s.UpdatedAt = now
	return s, nil
}

// Update changes the value and/or description of a setting.
func (s *SiteSetting) Update(value *string, description *string) {
	if value != nil {
		s.Value = *value
	}
	if description != nil {
		s.Description = *description
	}
	s.Touch()
}
