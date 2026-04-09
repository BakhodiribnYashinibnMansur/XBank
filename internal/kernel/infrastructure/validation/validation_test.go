package validation

import (
	"testing"
)

func TestStrongPassword_Valid(t *testing.T) {
	valid := []string{
		"Abcdef1!",
		"MyP@ssw0rd",
		"Str0ng#Pass",
		"Hello123$World",
	}
	for _, pw := range valid {
		if err := StrongPassword("password", pw); err != nil {
			t.Errorf("expected %q to be valid, got: %s", pw, err.Message)
		}
	}
}

func TestStrongPassword_TooShort(t *testing.T) {
	err := StrongPassword("password", "Ab1!")
	if err == nil {
		t.Error("expected error for short password")
	}
}

func TestStrongPassword_NoUppercase(t *testing.T) {
	err := StrongPassword("password", "abcdef1!")
	if err == nil {
		t.Error("expected error for missing uppercase")
	}
}

func TestStrongPassword_NoLowercase(t *testing.T) {
	err := StrongPassword("password", "ABCDEF1!")
	if err == nil {
		t.Error("expected error for missing lowercase")
	}
}

func TestStrongPassword_NoDigit(t *testing.T) {
	err := StrongPassword("password", "Abcdefgh!")
	if err == nil {
		t.Error("expected error for missing digit")
	}
}

func TestStrongPassword_NoSpecial(t *testing.T) {
	err := StrongPassword("password", "Abcdefg1")
	if err == nil {
		t.Error("expected error for missing special character")
	}
}
