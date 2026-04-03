package auth

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

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/user"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/auth"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/mock"
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

func setupAuthTest(t *testing.T) (*Service, *mock.UserRepository) {
	t.Helper()
	userRepo := mock.NewUserRepository()
	sessionRepo := mock.NewSessionRepository()
	jwtService := newTestJWTService(t)
	svc := NewService(userRepo, sessionRepo, jwtService, nil, nil, nil) // nil = no TOTP, no Redis in tests

	// Create a test user with hashed password
	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	userRepo.Create(context.Background(), &user.User{
		Email:     "ali@example.com",
		Password:  string(hashed),
		FirstName: "Ali",
		LastName:  "Valiyev",
	})

	return svc, userRepo
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
	if result.User.Email != "ali@example.com" {
		t.Errorf("expected: ali@example.com, got: %s", result.User.Email)
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

	// Login first to get a refresh token
	loginResult, _ := svc.Login(context.Background(), "ali@example.com", "password123", "TestAgent", "127.0.0.1")

	// Refresh with the token
	refreshResult, err := svc.Refresh(context.Background(), loginResult.RefreshToken, "TestAgent", "127.0.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if refreshResult.AccessToken == "" {
		t.Error("new access token should not be empty")
	}
	// New token should be different (rotation)
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

	// Refresh should fail after logout
	_, err = svc.Refresh(context.Background(), loginResult.RefreshToken, "TestAgent", "127.0.0.1")
	if err == nil {
		t.Error("refresh should fail after logout")
	}
}

// --- TOTP 2FA tests ---

func setupAuthTestWithTOTP(t *testing.T) (*Service, *mock.UserRepository) {
	t.Helper()
	userRepo := mock.NewUserRepository()
	sessionRepo := mock.NewSessionRepository()
	jwtService := newTestJWTService(t)
	totpService := infraAuth.NewTOTPService("XBank-Test")
	svc := NewService(userRepo, sessionRepo, jwtService, totpService, nil, nil)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	userRepo.Create(context.Background(), &user.User{
		Email:     "totp@example.com",
		Password:  string(hashed),
		FirstName: "TOTP",
		LastName:  "User",
	})

	return svc, userRepo
}

func TestSetupTOTP_Success(t *testing.T) {
	svc, userRepo := setupAuthTestWithTOTP(t)

	// Find user ID
	u, _ := userRepo.GetByEmail(context.Background(), "totp@example.com")

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

	// Secret should be saved but TOTP not yet enabled
	updated, _ := userRepo.GetByID(context.Background(), u.ID)
	if updated.TOTPSecret == "" {
		t.Error("secret should be saved to user")
	}
	if updated.TOTPEnabled {
		t.Error("TOTP should not be enabled before confirmation")
	}
}

func TestVerifyAndEnableTOTP_Success(t *testing.T) {
	svc, userRepo := setupAuthTestWithTOTP(t)
	u, _ := userRepo.GetByEmail(context.Background(), "totp@example.com")

	secret, _, _ := svc.SetupTOTP(context.Background(), u.ID)

	// Generate valid TOTP code
	code, _ := totp.GenerateCode(secret, time.Now())

	err := svc.VerifyAndEnableTOTP(context.Background(), u.ID, code)
	if err != nil {
		t.Fatal(err)
	}

	updated, _ := userRepo.GetByID(context.Background(), u.ID)
	if !updated.TOTPEnabled {
		t.Error("TOTP should be enabled after confirmation")
	}
}

func TestVerifyAndEnableTOTP_WrongCode(t *testing.T) {
	svc, userRepo := setupAuthTestWithTOTP(t)
	u, _ := userRepo.GetByEmail(context.Background(), "totp@example.com")

	svc.SetupTOTP(context.Background(), u.ID)

	err := svc.VerifyAndEnableTOTP(context.Background(), u.ID, "000000")
	if err == nil {
		t.Error("should reject wrong TOTP code")
	}
}

func TestLogin_TOTPRequired(t *testing.T) {
	svc, userRepo := setupAuthTestWithTOTP(t)
	u, _ := userRepo.GetByEmail(context.Background(), "totp@example.com")

	// Setup and enable TOTP
	secret, _, _ := svc.SetupTOTP(context.Background(), u.ID)
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.VerifyAndEnableTOTP(context.Background(), u.ID, code)

	// Login should return TOTPRequired
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
	svc, userRepo := setupAuthTestWithTOTP(t)
	u, _ := userRepo.GetByEmail(context.Background(), "totp@example.com")

	secret, _, _ := svc.SetupTOTP(context.Background(), u.ID)
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.VerifyAndEnableTOTP(context.Background(), u.ID, code)

	// Complete login with TOTP
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
	svc, userRepo := setupAuthTestWithTOTP(t)
	u, _ := userRepo.GetByEmail(context.Background(), "totp@example.com")

	secret, _, _ := svc.SetupTOTP(context.Background(), u.ID)
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.VerifyAndEnableTOTP(context.Background(), u.ID, code)

	_, err := svc.LoginWithTOTP(context.Background(), "totp@example.com", "000000", "Agent", "127.0.0.1")
	if err == nil {
		t.Error("should reject wrong TOTP code")
	}
}

func TestDisableTOTP_Success(t *testing.T) {
	svc, userRepo := setupAuthTestWithTOTP(t)
	u, _ := userRepo.GetByEmail(context.Background(), "totp@example.com")

	secret, _, _ := svc.SetupTOTP(context.Background(), u.ID)
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.VerifyAndEnableTOTP(context.Background(), u.ID, code)

	// Disable with correct password
	err := svc.DisableTOTP(context.Background(), u.ID, "password123")
	if err != nil {
		t.Fatal(err)
	}

	updated, _ := userRepo.GetByID(context.Background(), u.ID)
	if updated.TOTPEnabled {
		t.Error("TOTP should be disabled")
	}
}

func TestDisableTOTP_WrongPassword(t *testing.T) {
	svc, userRepo := setupAuthTestWithTOTP(t)
	u, _ := userRepo.GetByEmail(context.Background(), "totp@example.com")

	secret, _, _ := svc.SetupTOTP(context.Background(), u.ID)
	code, _ := totp.GenerateCode(secret, time.Now())
	svc.VerifyAndEnableTOTP(context.Background(), u.ID, code)

	err := svc.DisableTOTP(context.Background(), u.ID, "wrong-password")
	if err == nil {
		t.Error("should reject wrong password")
	}
}
