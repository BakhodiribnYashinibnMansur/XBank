package middleware

import (
	"bytes"
	"context"
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
	"strconv"
	"testing"
	"time"

	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/auth"
	infraCrypto "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/crypto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

func init() {
	logger.Init(true)
}

func newTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: apperror.ErrorHandler,
	})
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

	svc, err := infraAuth.NewJWTService(privPath, pubPath, "test", "test", 15*time.Minute, 24*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestRecoveryMiddleware(t *testing.T) {
	app := newTestApp()
	app.Use(RecoveryMiddleware())
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("test panic!")
	})

	req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Test xatolik: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Panic da 500 kutilgan, kelgan: %d", resp.StatusCode)
	}
}

func TestRequestIDMiddleware_GeneratesID(t *testing.T) {
	app := newTestApp()
	app.Use(RequestIDMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp, _ := app.Test(req)
	if resp.Header.Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header bo'sh bo'lmasligi kerak")
	}
}

func TestRequestIDMiddleware_UsesClientID(t *testing.T) {
	app := newTestApp()
	app.Use(RequestIDMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "my-custom-id")
	resp, _ := app.Test(req)
	if resp.Header.Get("X-Request-ID") != "my-custom-id" {
		t.Errorf("Client ID ishlatilishi kerak, kelgan: %s", resp.Header.Get("X-Request-ID"))
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	app := newTestApp()
	app.Use(RateLimitMiddleware(60, 1*time.Minute))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	var lastStatus int
	for i := 0; i < 65; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		resp, _ := app.Test(req)
		lastStatus = resp.StatusCode
	}
	if lastStatus != http.StatusTooManyRequests {
		t.Errorf("Rate limit 429 kutilgan, kelgan: %d", lastStatus)
	}
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	jwtService := newTestJWTService(t)

	app := newTestApp()
	app.Get("/protected", AuthMiddleware(jwtService), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Token siz 401 kutilgan, kelgan: %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	jwtService := newTestJWTService(t)

	app := newTestApp()
	app.Get("/protected", AuthMiddleware(jwtService), func(c *fiber.Ctx) error {
		userID := c.Locals("user_id").(string)
		email := c.Locals("email").(string)
		return c.JSON(fiber.Map{"user_id": userID, "email": email})
	})

	pair, _ := jwtService.GenerateTokenPair("user-123", "ali@example.com", "CUSTOMER", "0.0.0.0")
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Valid token bilan 200 kutilgan, kelgan: %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]string
	json.Unmarshal(body, &result)

	if result["user_id"] != "user-123" {
		t.Errorf("user_id kutilgan: user-123, kelgan: %s", result["user_id"])
	}
	if result["email"] != "ali@example.com" {
		t.Errorf("email kutilgan: ali@example.com, kelgan: %s", result["email"])
	}
}

// --- HMAC Middleware Tests ---

func newTestHMACSigner(t *testing.T) *infraCrypto.HMACSigner {
	t.Helper()
	secret, _ := infraCrypto.GenerateHMACSecret()
	signer, err := infraCrypto.NewHMACSigner(secret, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return signer
}

func TestHMACMiddleware_ValidSignature(t *testing.T) {
	signer := newTestHMACSigner(t)
	app := newTestApp()
	app.Post("/transfer", HMACMiddleware(signer), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	body := []byte(`{"amount":1000}`)
	ts := time.Now().Unix()
	sig := signer.Sign(ts, body)

	req, _ := http.NewRequest(http.MethodPost, "/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Signature-Timestamp", strconv.FormatInt(ts, 10))

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("valid HMAC should return 200, got %d", resp.StatusCode)
	}
}

func TestHMACMiddleware_MissingHeaders(t *testing.T) {
	signer := newTestHMACSigner(t)
	app := newTestApp()
	app.Post("/transfer", HMACMiddleware(signer), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodPost, "/transfer", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("missing HMAC headers should return 401, got %d", resp.StatusCode)
	}
}

func TestHMACMiddleware_TamperedBody(t *testing.T) {
	signer := newTestHMACSigner(t)
	app := newTestApp()
	app.Post("/transfer", HMACMiddleware(signer), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	body := []byte(`{"amount":1000}`)
	ts := time.Now().Unix()
	sig := signer.Sign(ts, body)

	tampered := []byte(`{"amount":999999}`)
	req, _ := http.NewRequest(http.MethodPost, "/transfer", bytes.NewReader(tampered))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Signature-Timestamp", strconv.FormatInt(ts, 10))

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("tampered body should return 401, got %d", resp.StatusCode)
	}
}

func TestHMACMiddleware_GET_SkipsCheck(t *testing.T) {
	signer := newTestHMACSigner(t)
	app := newTestApp()
	app.Get("/data", HMACMiddleware(signer), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/data", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET should skip HMAC check, got %d", resp.StatusCode)
	}
}

// --- Challenge Middleware Tests ---

type mockChallengeValidator struct {
	validTokens map[string]string // token → userID
}

func (v *mockChallengeValidator) ValidateToken(_ context.Context, token, userID string) error {
	if uid, ok := v.validTokens[token]; ok && uid == userID {
		return nil
	}
	return apperror.ErrChallengeTokenInvalid
}

func TestRequireChallenge_ValidToken(t *testing.T) {
	validator := &mockChallengeValidator{validTokens: map[string]string{"tok-abc": "user-1"}}
	app := newTestApp()
	app.Post("/transfer", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		return c.Next()
	}, RequireChallenge(validator), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodPost, "/transfer", nil)
	req.Header.Set("X-Challenge-Token", "tok-abc")
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("valid challenge token should return 200, got %d", resp.StatusCode)
	}
}

func TestRequireChallenge_MissingToken(t *testing.T) {
	validator := &mockChallengeValidator{validTokens: map[string]string{}}
	app := newTestApp()
	app.Post("/transfer", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		return c.Next()
	}, RequireChallenge(validator), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodPost, "/transfer", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("missing token should return 401, got %d", resp.StatusCode)
	}
}

func TestRequireChallenge_InvalidToken(t *testing.T) {
	validator := &mockChallengeValidator{validTokens: map[string]string{"tok-abc": "user-1"}}
	app := newTestApp()
	app.Post("/transfer", func(c *fiber.Ctx) error {
		c.Locals("user_id", "user-1")
		return c.Next()
	}, RequireChallenge(validator), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodPost, "/transfer", nil)
	req.Header.Set("X-Challenge-Token", "wrong-token")
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("invalid token should return 401, got %d", resp.StatusCode)
	}
}

func TestRequireChallenge_NilValidator_Skips(t *testing.T) {
	app := newTestApp()
	app.Post("/transfer", RequireChallenge(nil), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodPost, "/transfer", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("nil validator should skip, got %d", resp.StatusCode)
	}
}
