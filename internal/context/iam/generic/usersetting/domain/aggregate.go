package domain

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// UserSetting stores per-user preferences as key-value pairs.
// Natural uniqueness: (user_id, key).
type UserSetting struct {
	domain.AggregateRoot
	UserID string
	Key    string
	Value  string
}

func NewUserSetting(userID, key, value string) (*UserSetting, error) {
	if userID == "" {
		return nil, ErrEmptyUserID
	}
	if key == "" {
		return nil, ErrEmptyKey
	}
	now := time.Now()
	s := &UserSetting{UserID: userID, Key: key, Value: value}
	s.CreatedAt = now
	s.UpdatedAt = now
	return s, nil
}

func (s *UserSetting) UpdateValue(value string) {
	s.Value = value
	s.Touch()
}
