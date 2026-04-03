package auth

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestTOTP_GenerateSecret(t *testing.T) {
	svc := NewTOTPService("XBank")

	secret, url, err := svc.GenerateSecret("test@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if secret == "" {
		t.Error("secret should not be empty")
	}
	if url == "" {
		t.Error("URL should not be empty")
	}
	if len(secret) < 16 {
		t.Errorf("secret should be at least 16 chars, got %d", len(secret))
	}
}

func TestTOTP_Validate_ValidCode(t *testing.T) {
	svc := NewTOTPService("XBank")

	secret, _, _ := svc.GenerateSecret("test@example.com")

	// Generate a valid code from the secret
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if !svc.Validate(code, secret) {
		t.Error("valid TOTP code should pass validation")
	}
}

func TestTOTP_Validate_InvalidCode(t *testing.T) {
	svc := NewTOTPService("XBank")

	secret, _, _ := svc.GenerateSecret("test@example.com")

	if svc.Validate("000000", secret) {
		t.Error("invalid code should fail validation")
	}
}

func TestTOTP_Validate_WrongSecret(t *testing.T) {
	svc := NewTOTPService("XBank")

	secret1, _, _ := svc.GenerateSecret("user1@example.com")
	secret2, _, _ := svc.GenerateSecret("user2@example.com")

	code, _ := totp.GenerateCode(secret1, time.Now())

	if svc.Validate(code, secret2) {
		t.Error("code from different secret should fail")
	}
}
