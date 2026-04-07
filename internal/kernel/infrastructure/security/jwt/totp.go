package auth

import (
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPService - generates and validates TOTP codes.
//
// Uses HMAC-SHA1 with 30-second time steps and 6-digit codes.
// Compatible with Google Authenticator, Authy, and other TOTP apps.
type TOTPService struct {
	issuer string // displayed in authenticator app (e.g. "XBank")
}

// NewTOTPService creates a TOTP service for the given issuer name
func NewTOTPService(issuer string) *TOTPService {
	return &TOTPService{issuer: issuer}
}

// GenerateSecret - create a new TOTP secret for a user.
// Returns the base32-encoded secret and the otpauth:// URL for QR code.
func (s *TOTPService) GenerateSecret(email string) (secret, url string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

// Validate - check if a 6-digit TOTP code is valid for the given secret.
// Allows ±1 time step (30s) skew for clock drift tolerance.
func (s *TOTPService) Validate(code, secret string) bool {
	return totp.Validate(code, secret)
}
