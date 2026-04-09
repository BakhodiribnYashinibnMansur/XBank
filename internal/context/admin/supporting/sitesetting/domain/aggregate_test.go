package domain

import (
	"testing"
)

func TestNewSiteSetting_Success(t *testing.T) {
	s, err := NewSiteSetting("app_name", "XBank", SettingTypeGeneral, "Application name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Key != "app_name" {
		t.Errorf("Key expected app_name, got: %s", s.Key)
	}
	if s.Value != "XBank" {
		t.Errorf("Value expected XBank, got: %s", s.Value)
	}
	if s.SettingType != SettingTypeGeneral {
		t.Errorf("SettingType expected general, got: %s", s.SettingType)
	}
	if s.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewSiteSetting_EmptyKey(t *testing.T) {
	_, err := NewSiteSetting("", "value", SettingTypeGeneral, "")
	if err != ErrEmptyKey {
		t.Errorf("expected ErrEmptyKey, got: %v", err)
	}
}

func TestNewSiteSetting_EmptyValue(t *testing.T) {
	_, err := NewSiteSetting("key", "", SettingTypeGeneral, "")
	if err != ErrEmptyValue {
		t.Errorf("expected ErrEmptyValue, got: %v", err)
	}
}

func TestSiteSetting_Update(t *testing.T) {
	s, _ := NewSiteSetting("smtp_host", "smtp.xbank.uz", SettingTypeEmail, "SMTP server")

	newVal := "smtp2.xbank.uz"
	newDesc := "Updated SMTP server"
	s.Update(&newVal, &newDesc)

	if s.Value != "smtp2.xbank.uz" {
		t.Errorf("Value expected smtp2.xbank.uz, got: %s", s.Value)
	}
	if s.Description != "Updated SMTP server" {
		t.Errorf("Description mismatch, got: %s", s.Description)
	}
}

func TestSiteSetting_Update_NilFields(t *testing.T) {
	s, _ := NewSiteSetting("key", "original", SettingTypeGeneral, "original desc")

	s.Update(nil, nil)

	if s.Value != "original" {
		t.Errorf("Value should remain original when nil passed, got: %s", s.Value)
	}
	if s.Description != "original desc" {
		t.Errorf("Description should remain when nil passed, got: %s", s.Description)
	}
}

func TestSettingTypeConstants(t *testing.T) {
	tests := []struct {
		st   SettingType
		want string
	}{
		{SettingTypeGeneral, "general"},
		{SettingTypeEmail, "email"},
		{SettingTypeSecurity, "security"},
		{SettingTypePayment, "payment"},
	}

	for _, tt := range tests {
		if string(tt.st) != tt.want {
			t.Errorf("SettingType expected %s, got: %s", tt.want, tt.st)
		}
	}
}
