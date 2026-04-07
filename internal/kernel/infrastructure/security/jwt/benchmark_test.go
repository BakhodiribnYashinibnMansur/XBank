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

func newBenchJWTService(b *testing.B) *JWTService {
	b.Helper()
	dir := b.TempDir()
	privPath := filepath.Join(dir, "private.pem")
	pubPath := filepath.Join(dir, "public.pem")

	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privBytes, _ := x509.MarshalECPrivateKey(key)
	os.WriteFile(privPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}), 0600)
	pubBytes, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	os.WriteFile(pubPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes}), 0644)

	svc, err := NewJWTService(privPath, pubPath, "xbank", "xbank-api", 15*time.Minute, 24*time.Hour)
	if err != nil {
		b.Fatal(err)
	}
	return svc
}

func BenchmarkJWT_GenerateTokenPair(b *testing.B) {
	svc := newBenchJWTService(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GenerateTokenPair("user-123", "user@example.com", "CUSTOMER", "192.168.1.1")
	}
}

func BenchmarkJWT_ValidateToken(b *testing.B) {
	svc := newBenchJWTService(b)
	pair, _ := svc.GenerateTokenPair("user-123", "user@example.com", "CUSTOMER", "192.168.1.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.ValidateAccessToken(pair.AccessToken)
	}
}

func BenchmarkTOTP_Validate(b *testing.B) {
	totpService := NewTOTPService("XBank")
	secret, _, _ := totpService.GenerateSecret("user@example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		totpService.Validate("123456", secret) // will fail but measures computation
	}
}
