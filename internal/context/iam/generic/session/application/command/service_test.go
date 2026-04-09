package command

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/jwt"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/mock"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

func newTestJWTService(t *testing.T) *infraAuth.JWTService {
	t.Helper()
	dir := t.TempDir()
	privPath := filepath.Join(dir, "private.pem")
	pubPath := filepath.Join(dir, "public.pem")

	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privBytes, _ := x509.MarshalECPrivateKey(key)
	os.WriteFile(privPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}), 0600)
	pubBytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}), 0644)

	svc, _ := infraAuth.NewJWTService(privPath, pubPath, "test", "test", 15*time.Minute, 24*time.Hour)
	return svc
}

func setupAuthTest(t *testing.T) (*Service, *mock.MockUserAuthReader) {
	t.Helper()
	userAuth := mock.NewMockUserAuthReader()
	sessionRepo := mock.NewSessionRepository()
	jwtService := newTestJWTService(t)
	svc := NewService(userAuth, sessionRepo, jwtService, nil, nil, nil)

	// Create a test user with hashed password
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userAuth.CreateRaw("ali@example.com", string(hashed), "Ali", "Valiyev")

	return svc, userAuth
}

func TestLogin_Success(t *testing.T) {
	svc, _ := setupAuthTest(t)

	result, err := svc.Login(context.Background(), "ali@example.com", "password123", "TestAgent", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("access token should not be empty")
	}
	if result.RefreshToken == "" {
		t.Error("refresh token should not be empty")
	}
	if result.Email != "ali@example.com" {
		t.Errorf("expected: ali@example.com, got: %s", result.Email)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, _ := setupAuthTest(t)

	_, err := svc.Login(context.Background(), "ali@example.com", "wrongpassword", "TestAgent", "127.0.0.1")
	if err != ErrInvalidCredentials {
		t.Errorf("expected: %v, got: %v", ErrInvalidCredentials, err)
	}
}

func TestLogin_NonExistentUser(t *testing.T) {
	svc, _ := setupAuthTest(t)

	_, err := svc.Login(context.Background(), "nobody@example.com", "password123", "TestAgent", "127.0.0.1")
	if err != ErrInvalidCredentials {
		t.Errorf("expected: %v, got: %v", ErrInvalidCredentials, err)
	}
}

func TestRefresh_Success(t *testing.T) {
	svc, _ := setupAuthTest(t)

	loginResult, _ := svc.Login(context.Background(), "ali@example.com", "password123", "TestAgent", "127.0.0.1")

	refreshResult, err := svc.Refresh(context.Background(), loginResult.RefreshToken, "TestAgent", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refreshResult.AccessToken == "" {
		t.Error("new access token should not be empty")
	}
	if refreshResult.RefreshToken == loginResult.RefreshToken {
		t.Error("refresh token should be rotated (new one each time)")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	svc, _ := setupAuthTest(t)

	_, err := svc.Refresh(context.Background(), "invalid-token", "TestAgent", "127.0.0.1")
	if err == nil {
		t.Error("expected error for invalid refresh token")
	}
}

func TestLogout_Success(t *testing.T) {
	svc, _ := setupAuthTest(t)

	loginResult, _ := svc.Login(context.Background(), "ali@example.com", "password123", "TestAgent", "127.0.0.1")

	err := svc.Logout(context.Background(), loginResult.RefreshToken)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.Refresh(context.Background(), loginResult.RefreshToken, "TestAgent", "127.0.0.1")
	if err == nil {
		t.Error("refresh should fail after logout")
	}
}

// --- TOTP 2FA tests ---

func setupAuthTestWithTOTP(t *testing.T) (*Service, *mock.MockUserAuthReader) {
	t.Helper()
	userAuth := mock.NewMockUserAuthReader()
	sessionRepo := mock.NewSessionRepository()
	jwtService := newTestJWTService(t)
	totpService := infraAuth.NewTOTPService("XBank-Test")
	svc := NewService(userAuth, sessionRepo, jwtService, totpService, nil, nil)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	userAuth.CreateRaw("totp@example.com", string(hashed), "TOTP", "User")

	return svc, userAuth
}

func TestSetupTOTP_Success(t *testing.T) {
	svc, userAuth := setupAuthTestWithTOTP(t)

	u, _ := userAuth.GetInternalByEmail(context.Background(), "totp@example.com")

	secret, url, err := svc.SetupTOTP(context.Background(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if secret == "" {
		t.Error("secret should not be empty")
	}
	if url == "" {
		t.Error("URL should not be empty")
	}

	updated, _ := userAuth.GetInternalByID(context.Background(), u.ID)
	if updated.TOTPSecret == "" {
		t.Error("secret should be saved to user")
	}
	if updated.TOTPEnabled {
		t.Error("TOTP should not be enabled before confirmation")
	}
}

func TestVerifyAndEnableTOTP_Success(t *testing.T) {
	svc, userAuth := setupAuthTestWithTOTP(t)
	u, _ := userAuth.GetInternalByEmail(context.Background(), "totp@example.com")

	secret, _, _ := svc.SetupTOTP(context.Background(), u.ID)

	code, _ := totp.GenerateCode(secret, time.Now())

	err := svc.VerifyAndEnableTOTP(context.Background(), u.ID, code)
	if err != nil {
		t.Fatal(err)
	}

	updated, _ := userAuth.GetInternalByID(context.Background(), u.ID)
	if !updated.TOTPEnabled {
		t.Error("TOTP should be enabled after confirmation")
	}
}

func TestVerifyAndEnableTOTP_WrongCode(t *testing.T) {
	svc, userAuth := setupAuthTestWithTOTP(t)
	u, _ := userAuth.GetInternalByEmail(context.Background(), "totp@example.com")

	svc.SetupTOTP(context.Background(), u.ID)

	err := svc.VerifyAndEnableTOTP(context.Background(), u.ID, "000000")
	if err == nil {
		t.Error("should reject wrong TOTP code")
	}
}

func TestLogin_TOTPRequired(t *testing.T) {
	svc, userAuth := setupAuthTestWithTOTP(t)
	u, _ := userAuth.GetInternalByEmail(context.Background(), "totp@example.com")

	secret, _, _ := svc.SetupTOTP(context.Background(), u.ID)
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.VerifyAndEnableTOTP(context.Background(), u.ID, code)

	result, err := svc.Login(context.Background(), "totp@example.com", "password123", "Agent", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if !result.TOTPRequired {
		t.Error("login should require TOTP")
	}
	if result.AccessToken != "" {
		t.Error("tokens should NOT be issued before TOTP verification")
	}
}

func TestLoginWithTOTP_Success(t *testing.T) {
	svc, userAuth := setupAuthTestWithTOTP(t)
	u, _ := userAuth.GetInternalByEmail(context.Background(), "totp@example.com")

	secret, _, _ := svc.SetupTOTP(context.Background(), u.ID)
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.VerifyAndEnableTOTP(context.Background(), u.ID, code)

	code2, _ := totp.GenerateCode(secret, time.Now())
	result, err := svc.LoginWithTOTP(context.Background(), "totp@example.com", code2, "Agent", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	if result.AccessToken == "" {
		t.Error("access token should be issued after TOTP")
	}
}

func TestLoginWithTOTP_WrongCode(t *testing.T) {
	svc, userAuth := setupAuthTestWithTOTP(t)
	u, _ := userAuth.GetInternalByEmail(context.Background(), "totp@example.com")

	secret, _, _ := svc.SetupTOTP(context.Background(), u.ID)
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.VerifyAndEnableTOTP(context.Background(), u.ID, code)

	_, err := svc.LoginWithTOTP(context.Background(), "totp@example.com", "000000", "Agent", "127.0.0.1")
	if err == nil {
		t.Error("should reject wrong TOTP code")
	}
}

func TestDisableTOTP_Success(t *testing.T) {
	svc, userAuth := setupAuthTestWithTOTP(t)
	u, _ := userAuth.GetInternalByEmail(context.Background(), "totp@example.com")

	secret, _, _ := svc.SetupTOTP(context.Background(), u.ID)
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.VerifyAndEnableTOTP(context.Background(), u.ID, code)

	err := svc.DisableTOTP(context.Background(), u.ID, "password123")
	if err != nil {
		t.Fatal(err)
	}

	updated, _ := userAuth.GetInternalByID(context.Background(), u.ID)
	if updated.TOTPEnabled {
		t.Error("TOTP should be disabled")
	}
}

func TestDisableTOTP_WrongPassword(t *testing.T) {
	svc, userAuth := setupAuthTestWithTOTP(t)
	u, _ := userAuth.GetInternalByEmail(context.Background(), "totp@example.com")

	secret, _, _ := svc.SetupTOTP(context.Background(), u.ID)
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.VerifyAndEnableTOTP(context.Background(), u.ID, code)

	err := svc.DisableTOTP(context.Background(), u.ID, "wrong-password")
	if err == nil {
		t.Error("should reject wrong password")
	}
}
