package domain

import (
	"testing"
)

func TestNewBeneficiary_Success(t *testing.T) {
	b, err := NewBeneficiary("user-1", "John Doe", "1234567890123456", "XBank", "XBNK", "UZS", TypeInternal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.UserID != "user-1" {
		t.Errorf("UserID expected user-1, got: %s", b.UserID)
	}
	if b.Name != "John Doe" {
		t.Errorf("Name expected John Doe, got: %s", b.Name)
	}
	if b.AccountNumber != "1234567890123456" {
		t.Errorf("AccountNumber expected 1234567890123456, got: %s", b.AccountNumber)
	}
	if b.BenType != TypeInternal {
		t.Errorf("BenType expected INTERNAL, got: %s", b.BenType)
	}
	if b.Verified {
		t.Error("new beneficiary should not be verified")
	}
	if b.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewBeneficiary_MissingFields(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		benName       string
		accountNumber string
	}{
		{"missing user_id", "", "John", "1234"},
		{"missing name", "user-1", "", "1234"},
		{"missing account_number", "user-1", "John", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewBeneficiary(tt.userID, tt.benName, tt.accountNumber, "XBank", "", "UZS", TypeInternal)
			if err == nil {
				t.Error("expected error for missing required field")
			}
		})
	}
}

func TestBeneficiaryTypes(t *testing.T) {
	tests := []struct {
		benType Type
		want    string
	}{
		{TypeInternal, "INTERNAL"},
		{TypeExternal, "EXTERNAL"},
		{TypeInternational, "INTERNATIONAL"},
	}

	for _, tt := range tests {
		t.Run(string(tt.benType), func(t *testing.T) {
			b, err := NewBeneficiary("user-1", "Name", "1234", "Bank", "", "UZS", tt.benType)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(b.BenType) != tt.want {
				t.Errorf("BenType expected %s, got: %s", tt.want, b.BenType)
			}
		})
	}
}
