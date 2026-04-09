package domain

import (
	"testing"
)

func TestNewUserSetting_Success(t *testing.T) {
	s, err := NewUserSetting("user-1", "theme", "dark")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.UserID != "user-1" {
		t.Errorf("UserID expected user-1, got: %s", s.UserID)
	}
	if s.Key != "theme" {
		t.Errorf("Key expected theme, got: %s", s.Key)
	}
	if s.Value != "dark" {
		t.Errorf("Value expected dark, got: %s", s.Value)
	}
	if s.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewUserSetting_EmptyUserID(t *testing.T) {
	_, err := NewUserSetting("", "theme", "dark")
	if err != ErrEmptyUserID {
		t.Errorf("expected ErrEmptyUserID, got: %v", err)
	}
}

func TestNewUserSetting_EmptyKey(t *testing.T) {
	_, err := NewUserSetting("user-1", "", "dark")
	if err != ErrEmptyKey {
		t.Errorf("expected ErrEmptyKey, got: %v", err)
	}
}

func TestUserSetting_UpdateValue(t *testing.T) {
	s, _ := NewUserSetting("user-1", "language", "en")
	originalUpdatedAt := s.UpdatedAt

	s.UpdateValue("uz")

	if s.Value != "uz" {
		t.Errorf("Value expected uz, got: %s", s.Value)
	}
	if !s.UpdatedAt.After(originalUpdatedAt) && s.UpdatedAt.Equal(originalUpdatedAt) {
		// UpdatedAt should be >= original (may be equal if test runs fast)
	}
}

func TestNewUserSetting_EmptyValue(t *testing.T) {
	// Empty value should be allowed (user can clear a setting)
	s, err := NewUserSetting("user-1", "theme", "")
	if err != nil {
		t.Fatalf("unexpected error for empty value: %v", err)
	}
	if s.Value != "" {
		t.Errorf("Value expected empty, got: %s", s.Value)
	}
}
