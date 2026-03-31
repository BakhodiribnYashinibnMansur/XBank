package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// generateTestKeys - test uchun vaqtinchalik PEM fayllar yaratadi
func generateTestKeys(t *testing.T) (privatePath, publicPath string) {
	t.Helper()

	dir := t.TempDir()
	privatePath = filepath.Join(dir, "private.pem")
	publicPath = filepath.Join(dir, "public.pem")

	// ECDSA P-256 key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	// Private key → PEM
	privBytes, _ := x509.MarshalECPrivateKey(privateKey)
	privPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	os.WriteFile(privatePath, privPem, 0600)

	// Public key → PEM
	pubBytes, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	os.WriteFile(publicPath, pubPem, 0644)

	return privatePath, publicPath
}

func newTestJWTService(t *testing.T) *JWTService {
	t.Helper()
	privPath, pubPath := generateTestKeys(t)
	svc, err := NewJWTService(privPath, pubPath, "test-issuer", "test-audience", 15*time.Minute, 30*24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestGenerateAndValidateToken(t *testing.T) {
	svc := newTestJWTService(t)

	pair, err := svc.GenerateTokenPair("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("Token yaratishda xatolik: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("Access token bo'sh bo'lmasligi kerak")
	}
	if pair.RefreshToken == "" {
		t.Error("Refresh token bo'sh bo'lmasligi kerak")
	}

	// Validate
	claims, err := svc.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("Token validate xatolik: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("UserID kutilgan: user-123, kelgan: %s", claims.UserID)
	}
	if claims.Email != "test@example.com" {
		t.Errorf("Email kutilgan: test@example.com, kelgan: %s", claims.Email)
	}
	if claims.Issuer != "test-issuer" {
		t.Errorf("Issuer kutilgan: test-issuer, kelgan: %s", claims.Issuer)
	}
}

func TestValidateExpiredToken(t *testing.T) {
	privPath, pubPath := generateTestKeys(t)
	svc, _ := NewJWTService(privPath, pubPath, "test", "test", 1*time.Nanosecond, 24*time.Hour)

	pair, _ := svc.GenerateTokenPair("user-123", "test@example.com")
	time.Sleep(2 * time.Millisecond)

	_, err := svc.ValidateAccessToken(pair.AccessToken)
	if err != ErrExpiredToken {
		t.Errorf("Kutilgan: %v, kelgan: %v", ErrExpiredToken, err)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	svc := newTestJWTService(t)

	_, err := svc.ValidateAccessToken("invalid.token.string")
	if err != ErrInvalidToken {
		t.Errorf("Kutilgan: %v, kelgan: %v", ErrInvalidToken, err)
	}
}

func TestValidateTokenWithWrongKey(t *testing.T) {
	// Bitta key bilan sign, boshqa key bilan verify → xatolik
	svc1 := newTestJWTService(t)
	svc2 := newTestJWTService(t) // boshqa key pair

	pair, _ := svc1.GenerateTokenPair("user-123", "test@example.com")

	_, err := svc2.ValidateAccessToken(pair.AccessToken)
	if err != ErrInvalidToken {
		t.Errorf("Boshqa key bilan verify xatolik bo'lishi kerak, kelgan: %v", err)
	}
}

func TestHashToken(t *testing.T) {
	hash1 := HashToken("my-refresh-token")
	hash2 := HashToken("my-refresh-token")
	hash3 := HashToken("different-token")

	if hash1 != hash2 {
		t.Error("Bir xil token = bir xil hash")
	}
	if hash1 == hash3 {
		t.Error("Turli tokenlar = turli hash")
	}
}
