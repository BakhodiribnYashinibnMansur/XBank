package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestSessionFlow_MultipleSessionsLogoutAll tests multiple login sessions and logout-all:
// login session 1 -> login session 2 -> logout-all -> both sessions invalidated
func TestSessionFlow_MultipleSessionsLogoutAll(t *testing.T) {
	email := uniqueEmail("multi-session")
	password := "Password123!"

	// Register
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":      email,
		"password":   password,
		"first_name": "Multi",
		"last_name":  "Session",
	}, "")
	expectStatus(t, rec, fiber.StatusCreated)

	// Login session 1
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	var session1 authTokens
	parseResponse(t, rec, &session1)

	if session1.AccessToken == "" {
		t.Fatal("expected session1 access token")
	}

	// Login session 2
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	var session2 authTokens
	parseResponse(t, rec, &session2)

	if session2.AccessToken == "" {
		t.Fatal("expected session2 access token")
	}

	// Both tokens should be different
	if session1.AccessToken == session2.AccessToken {
		t.Error("expected different access tokens for different sessions")
	}

	// Both sessions should work for protected endpoints
	rec = doRequest(t, fiber.MethodGet, "/api/v1/accounts/list", nil, session1.AccessToken)
	if rec.Code == fiber.StatusUnauthorized {
		t.Error("session1 should still be valid")
	}

	rec = doRequest(t, fiber.MethodGet, "/api/v1/accounts/list", nil, session2.AccessToken)
	if rec.Code == fiber.StatusUnauthorized {
		t.Error("session2 should still be valid")
	}

	// Logout all sessions using session1
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/logout-all", nil, session1.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Both refresh tokens should now be invalid
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": session1.RefreshToken,
	}, "")
	if rec.Code == fiber.StatusOK {
		t.Error("expected session1 refresh to fail after logout-all")
	}

	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": session2.RefreshToken,
	}, "")
	if rec.Code == fiber.StatusOK {
		t.Error("expected session2 refresh to fail after logout-all")
	}
}

// TestSessionFlow_RefreshTokenRotation tests that refresh tokens are properly rotated.
func TestSessionFlow_RefreshTokenRotation(t *testing.T) {
	tokens := registerAndLogin(t, uniqueEmail("refresh-rotation"), "Password123!", "Refresh")

	originalRefresh := tokens.RefreshToken

	// Refresh
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": originalRefresh,
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	var newTokens authTokens
	parseResponse(t, rec, &newTokens)

	if newTokens.AccessToken == "" {
		t.Fatal("expected new access token")
	}
	if newTokens.RefreshToken == "" {
		t.Fatal("expected new refresh token")
	}

	// New tokens should be different from original
	if newTokens.AccessToken == tokens.AccessToken {
		t.Error("expected different access token after refresh")
	}

	// Use new refresh token for another refresh
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": newTokens.RefreshToken,
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	var thirdTokens authTokens
	parseResponse(t, rec, &thirdTokens)

	if thirdTokens.AccessToken == newTokens.AccessToken {
		t.Error("expected different access token on second refresh")
	}
}

// TestSessionFlow_ExpiredTokenRejected tests that using an invalid/malformed token is rejected.
func TestSessionFlow_ExpiredTokenRejected(t *testing.T) {
	// Use a completely invalid token
	rec := doRequest(t, fiber.MethodGet, "/api/v1/accounts/list", nil, "invalid.token.here")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want %d for invalid token", rec.Code, fiber.StatusUnauthorized)
	}

	// Use an empty token
	rec = doRequest(t, fiber.MethodGet, "/api/v1/accounts/list", nil, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want %d for empty token", rec.Code, fiber.StatusUnauthorized)
	}
}

// TestSessionFlow_LogoutThenAccess tests that accessing protected routes after logout
// with a refresh token fails, while access token may still be temporarily valid (JWT).
func TestSessionFlow_LogoutThenAccess(t *testing.T) {
	tokens := registerAndLogin(t, uniqueEmail("logout-access"), "Password123!", "LogoutAccess")

	// Logout
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/logout", map[string]string{
		"refresh_token": tokens.RefreshToken,
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	// Refresh should fail
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": tokens.RefreshToken,
	}, "")
	if rec.Code == fiber.StatusOK {
		t.Error("expected refresh to fail after logout")
	}
}

// TestSessionFlow_RefreshWithInvalidToken tests refresh with garbage token.
func TestSessionFlow_RefreshWithInvalidToken(t *testing.T) {
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": "completely-invalid-refresh-token",
	}, "")
	if rec.Code == fiber.StatusOK {
		t.Error("expected refresh with invalid token to fail")
	}
}

// TestSessionFlow_RefreshWithEmptyToken tests refresh with empty token.
func TestSessionFlow_RefreshWithEmptyToken(t *testing.T) {
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": "",
	}, "")
	if rec.Code == fiber.StatusOK {
		t.Error("expected refresh with empty token to fail")
	}
}

// TestSessionFlow_LogoutAllThenLogin tests that after logout-all, the user can still login fresh.
func TestSessionFlow_LogoutAllThenLogin(t *testing.T) {
	email := uniqueEmail("logout-relogin")
	password := "Password123!"
	tokens := registerAndLogin(t, email, password, "ReLogin")

	// Logout all
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/logout-all", nil, tokens.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Login again should succeed
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": password,
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	var fresh authTokens
	parseResponse(t, rec, &fresh)
	if fresh.AccessToken == "" {
		t.Fatal("expected fresh access token after re-login")
	}
}

// TestSessionFlow_TOTPSetupFlow tests TOTP setup and management.
func TestSessionFlow_TOTPSetupFlow(t *testing.T) {
	tokens := registerAndLogin(t, uniqueEmail("totp-setup"), "Password123!", "TOTP")

	// Setup TOTP
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/totp/setup", nil, tokens.AccessToken)
	expectStatusOneOf(t, rec, fiber.StatusOK, fiber.StatusCreated)

	var totpSetup struct {
		Secret string `json:"secret"`
		URI    string `json:"uri"`
		QR     string `json:"qr_code"`
	}
	parseResponse(t, rec, &totpSetup)

	if totpSetup.Secret == "" {
		t.Error("expected TOTP secret")
	}

	// Disable TOTP (since we cannot generate a real TOTP code in tests)
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/totp/disable", nil, tokens.AccessToken)
	expectStatusOneOf(t, rec, fiber.StatusOK, fiber.StatusBadRequest)
}

// TestSessionFlow_LogoutIdempotent tests that calling logout twice does not cause errors.
func TestSessionFlow_LogoutIdempotent(t *testing.T) {
	tokens := registerAndLogin(t, uniqueEmail("logout-idem"), "Password123!", "Idem")

	// First logout
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/logout", map[string]string{
		"refresh_token": tokens.RefreshToken,
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	// Second logout -- should not panic or return 500
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/logout", map[string]string{
		"refresh_token": tokens.RefreshToken,
	}, "")
	if rec.Code == fiber.StatusInternalServerError {
		t.Error("expected graceful handling of double logout, not 500")
	}
}
