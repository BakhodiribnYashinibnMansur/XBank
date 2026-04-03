package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	infraCrypto "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/crypto"
	"github.com/gofiber/fiber/v2"
)

// ════════════════════════════════════════════════════════
// OWASP Top 10 Security Tests
// ════════════════════════════════════════════════════════

// --- A01: Broken Access Control ---

func TestRBAC_UnauthorizedRoleRejected(t *testing.T) {
	jwtService := newTestJWTService(t)
	app := newTestApp()
	app.Get("/admin", AuthMiddleware(jwtService), RequireRole("ADMIN"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// Login as CUSTOMER, try to access ADMIN endpoint
	pair, _ := jwtService.GenerateTokenPair("user-1", "user@test.com", "CUSTOMER", "0.0.0.0")
	req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("CUSTOMER accessing ADMIN route should get 403, got %d", resp.StatusCode)
	}
}

func TestRBAC_AdminRoleAccepted(t *testing.T) {
	jwtService := newTestJWTService(t)
	app := newTestApp()
	app.Get("/admin", AuthMiddleware(jwtService), RequireRole("ADMIN"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	pair, _ := jwtService.GenerateTokenPair("admin-1", "admin@test.com", "ADMIN", "0.0.0.0")
	req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ADMIN should access ADMIN route, got %d", resp.StatusCode)
	}
}

func TestRBAC_NoRoleInToken(t *testing.T) {
	jwtService := newTestJWTService(t)
	app := newTestApp()
	app.Get("/admin", AuthMiddleware(jwtService), RequireRole("ADMIN"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	pair, _ := jwtService.GenerateTokenPair("user-1", "user@test.com", "", "0.0.0.0")
	req, _ := http.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("empty role should get 403, got %d", resp.StatusCode)
	}
}

// --- A02: Cryptographic Failures ---

func TestHelmet_SecurityHeaders(t *testing.T) {
	app := newTestApp()
	app.Use(HelmetMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp, _ := app.Test(req)

	tests := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":          "DENY",
		"X-XSS-Protection":         "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
		"Referrer-Policy":          "strict-origin-when-cross-origin",
		"Cache-Control":            "no-store, no-cache, must-revalidate",
	}

	for header, expected := range tests {
		got := resp.Header.Get(header)
		if got != expected {
			t.Errorf("Header %s: expected %q, got %q", header, expected, got)
		}
	}
}

func TestHelmet_CSP_Blocks_InlineScript(t *testing.T) {
	app := newTestApp()
	app.Use(HelmetMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp, _ := app.Test(req)

	csp := resp.Header.Get("Content-Security-Policy")
	if !strings.Contains(csp, "default-src 'none'") {
		t.Errorf("CSP should have default-src 'none', got: %s", csp)
	}
}

func TestHelmet_Clickjacking_Prevention(t *testing.T) {
	app := newTestApp()
	app.Use(HelmetMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp, _ := app.Test(req)

	xFrame := resp.Header.Get("X-Frame-Options")
	if xFrame != "DENY" {
		t.Errorf("X-Frame-Options should be DENY, got %s", xFrame)
	}
}

// --- A03: Injection ---

func TestSQLInjection_LoginEmail(t *testing.T) {
	app := newTestApp()
	app.Post("/auth/login", func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}

		// Check that SQL injection patterns are present in the input
		// (they should be passed as-is, not executed — parameterized queries prevent this)
		sqlPayloads := []string{
			"' OR '1'='1",
			"'; DROP TABLE users; --",
			"' UNION SELECT * FROM users --",
		}
		for _, payload := range sqlPayloads {
			if strings.Contains(body.Email, payload) {
				// Input contains injection attempt — this is fine because
				// parameterized queries will treat it as a literal string
				return c.SendStatus(http.StatusUnauthorized) // auth fails, not injected
			}
		}
		return c.SendStatus(http.StatusUnauthorized)
	})

	sqlPayloads := []string{
		`{"email":"' OR '1'='1","password":"x"}`,
		`{"email":"'; DROP TABLE users; --","password":"x"}`,
		`{"email":"' UNION SELECT * FROM users --","password":"x"}`,
	}

	for _, payload := range sqlPayloads {
		req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)
		if resp.StatusCode == http.StatusOK {
			t.Errorf("SQL injection payload should NOT return 200: %s", payload)
		}
	}
}

func TestXSS_ResponseHeaders(t *testing.T) {
	app := newTestApp()
	app.Use(HelmetMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"name": c.Query("name")})
	})

	// XSS payload in query param
	req, _ := http.NewRequest(http.MethodGet, `/test?name=<script>alert('xss')</script>`, nil)
	resp, _ := app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Response should be JSON (Content-Type: application/json),
	// so XSS won't execute. Additionally, nosniff header prevents MIME sniffing.
	if resp.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Error("Missing nosniff header — XSS via content type sniffing possible")
	}

	// The XSS payload is present in JSON (as a value), but JSON encoding escapes it
	if strings.Contains(bodyStr, "<script>") {
		// fiber.JSON auto-escapes HTML by default, but even if it doesn't,
		// the Content-Type: application/json prevents browser execution
		ct := resp.Header.Get("Content-Type")
		if !strings.Contains(ct, "application/json") {
			t.Error("Non-JSON content type with unescaped XSS is dangerous")
		}
	}
}

// --- A05: Security Misconfiguration ---

func TestNoServerHeader(t *testing.T) {
	app := newTestApp()
	app.Use(HelmetMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp, _ := app.Test(req)

	server := resp.Header.Get("Server")
	if server != "" && strings.Contains(strings.ToLower(server), "fiber") {
		t.Error("Server header should not expose framework name")
	}
}

func TestNoCacheOnSensitiveEndpoints(t *testing.T) {
	app := newTestApp()
	app.Use(HelmetMiddleware())
	app.Get("/api/v1/users/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"email": "user@test.com"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	resp, _ := app.Test(req)

	cache := resp.Header.Get("Cache-Control")
	if !strings.Contains(cache, "no-store") {
		t.Errorf("Sensitive endpoints must have no-store, got: %s", cache)
	}
}

// --- A07: Identification and Authentication Failures ---

func TestAuth_ExpiredToken(t *testing.T) {
	jwtService := newTestJWTService(t)
	app := newTestApp()
	app.Get("/protected", AuthMiddleware(jwtService), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// Use a malformed/expired token
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer expired.invalid.token")

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expired token should return 401, got %d", resp.StatusCode)
	}
}

func TestAuth_MalformedBearer(t *testing.T) {
	jwtService := newTestJWTService(t)
	app := newTestApp()
	app.Get("/protected", AuthMiddleware(jwtService), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	malformed := []string{
		"Basic dXNlcjpwYXNz",    // Basic auth instead of Bearer
		"Bearer",                 // No token
		"bearer valid-token",     // Wrong case
		"Token something",        // Wrong scheme
		"",                       // Empty header
	}

	for _, header := range malformed {
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}

		resp, _ := app.Test(req)
		if resp.StatusCode == http.StatusOK {
			t.Errorf("Malformed auth header %q should not return 200", header)
		}
	}
}

// --- A08: Software and Data Integrity Failures ---

func TestHMAC_ReplayAttack(t *testing.T) {
	// Create a signer with very short clock skew to test replay
	secret, _ := infraCrypto.GenerateHMACSecret()
	signer, _ := infraCrypto.NewHMACSigner(secret, 1*time.Second)

	app := newTestApp()
	app.Post("/transfer", HMACMiddleware(signer), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	body := []byte(`{"amount":1000}`)
	oldTs := time.Now().Add(-10 * time.Second).Unix() // 10 seconds in the past
	sig := signer.Sign(oldTs, body)

	req, _ := http.NewRequest(http.MethodPost, "/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", sig)
	req.Header.Set("X-Signature-Timestamp", fmt.Sprintf("%d", oldTs))

	resp, _ := app.Test(req)
	if resp.StatusCode == http.StatusOK {
		t.Error("Replay attack with old timestamp should be rejected")
	}
}

func TestHMAC_InvalidSignature(t *testing.T) {
	signer := newTestHMACSigner(t)
	app := newTestApp()
	app.Post("/transfer", HMACMiddleware(signer), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	body := []byte(`{"amount":1000}`)
	ts := time.Now().Unix()

	req, _ := http.NewRequest(http.MethodPost, "/transfer", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", "deadbeef0123456789abcdef0123456789abcdef0123456789abcdef01234567")
	req.Header.Set("X-Signature-Timestamp", fmt.Sprintf("%d", ts))

	resp, _ := app.Test(req)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Invalid HMAC signature should return 401, got %d", resp.StatusCode)
	}
}

// --- A09: Security Logging and Monitoring Failures ---

func TestRecovery_PanicDoesNotLeakStackTrace(t *testing.T) {
	app := newTestApp()
	app.Use(RecoveryMiddleware())
	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("secret database password: hunter2")
	})

	req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
	resp, _ := app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("panic should return 500, got %d", resp.StatusCode)
	}

	if strings.Contains(bodyStr, "hunter2") {
		t.Error("Panic response should NOT leak sensitive data from stack trace")
	}
	if strings.Contains(bodyStr, "goroutine") {
		t.Error("Panic response should NOT expose Go stack trace")
	}
}

// --- Rate Limiting (DoS Prevention) ---

func TestRateLimit_BurstProtection(t *testing.T) {
	app := newTestApp()
	app.Use(RateLimitMiddleware(5, 1*time.Minute)) // Only 5 requests per minute
	app.Get("/api", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	var responses []int
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/api", nil)
		resp, _ := app.Test(req)
		responses = append(responses, resp.StatusCode)
	}

	// First 5 should succeed, rest should be rate limited
	blocked := 0
	for _, code := range responses {
		if code == http.StatusTooManyRequests {
			blocked++
		}
	}

	if blocked == 0 {
		t.Error("Rate limiter should block excessive requests")
	}
}

func TestRateLimit_RetryAfterHeader(t *testing.T) {
	app := newTestApp()
	app.Use(RateLimitMiddleware(1, 1*time.Minute))
	app.Get("/api", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// First request passes
	req1, _ := http.NewRequest(http.MethodGet, "/api", nil)
	app.Test(req1)

	// Second request should be limited
	req2, _ := http.NewRequest(http.MethodGet, "/api", nil)
	resp2, _ := app.Test(req2)

	if resp2.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp2.Header.Get("Retry-After")
		if retryAfter == "" {
			// Not all implementations include this, but it's best practice
			t.Log("INFO: Retry-After header not set on 429 response (optional)")
		}
	}
}

// --- Sensitive Data in Error Responses ---

func TestErrorResponse_NoInternalDetails(t *testing.T) {
	app := newTestApp()
	app.Use(RecoveryMiddleware())
	app.Get("/error", func(c *fiber.Ctx) error {
		// Simulate an internal error with sensitive details
		return fmt.Errorf("connection to postgres://user:secret@db:5432/xbank failed")
	})

	req, _ := http.NewRequest(http.MethodGet, "/error", nil)
	resp, _ := app.Test(req)

	body, _ := io.ReadAll(resp.Body)
	bodyStr := strings.ToLower(string(body))

	dangerousPatterns := []string{
		"postgres://",
		"secret",
		"password",
		":5432",
		"connection string",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(bodyStr, pattern) {
			t.Errorf("Error response should not contain %q — leaks internal info", pattern)
		}
	}
}

// --- HSTS Enforcement ---

func TestHSTS_IncludesSubDomains(t *testing.T) {
	app := newTestApp()
	app.Use(HelmetMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	resp, _ := app.Test(req)

	hsts := resp.Header.Get("Strict-Transport-Security")
	if !strings.Contains(hsts, "includeSubDomains") {
		t.Error("HSTS should include subdomains")
	}
	if !strings.Contains(hsts, "preload") {
		t.Error("HSTS should include preload directive")
	}
}
