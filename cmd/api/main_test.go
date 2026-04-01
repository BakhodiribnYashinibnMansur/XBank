package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/auth"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/config"
	router "github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/handler"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
)

func init() {
	logger.Init(true)
}

func testConfig() *config.Config {
	return &config.Config{
		App:       config.AppConfig{Name: "XBank Test", Port: 3000},
		RateLimit: config.RateLimitConfig{MaxRequests: 60, WindowMinutes: 1},
		CORS:      config.CORSConfig{AllowedOrigins: []string{"http://localhost:3000"}},
	}
}

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

func TestHealthEndpoint(t *testing.T) {
	cfg := testConfig()
	jwtService := newTestJWTService(t)
	app := router.NewRouter(&handler.UserHandler{}, &handler.AuthHandler{}, &handler.AccountHandler{}, &handler.TransferHandler{}, &handler.CardHandler{}, &handler.BeneficiaryHandler{}, &handler.ExchangeHandler{}, &handler.KYCHandler{}, &handler.FraudHandler{}, &handler.HealthHandler{}, jwtService, nil, cfg)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test requestda xatolik: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status code kutilgan: %d, kelgan: %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	if result["status"] != "ok" {
		t.Errorf("Status kutilgan: 'ok', kelgan: '%s'", result["status"])
	}
}

func TestProtectedRouteWithoutToken(t *testing.T) {
	cfg := testConfig()
	jwtService := newTestJWTService(t)
	app := router.NewRouter(&handler.UserHandler{}, &handler.AuthHandler{}, &handler.AccountHandler{}, &handler.TransferHandler{}, &handler.CardHandler{}, &handler.BeneficiaryHandler{}, &handler.ExchangeHandler{}, &handler.KYCHandler{}, &handler.FraudHandler{}, &handler.HealthHandler{}, jwtService, nil, cfg)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/get?id=123", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test requestda xatolik: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Token siz 401 kutilgan, kelgan: %d", resp.StatusCode)
	}
}
