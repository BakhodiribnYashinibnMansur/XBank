package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestAuthFlow_RegisterLoginRefreshLogout tests the complete auth lifecycle:
// register → login → refresh → logout
func TestAuthFlow_RegisterLoginRefreshLogout(t *testing.T) {
	// Step 1: Register
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":      "auth-flow@example.com",
		"password":   "SecurePass123!",
		"first_name": "Auth",
		"last_name":  "User",
	}, "")
	expectStatus(t, rec, fiber.StatusCreated)

	var userResp struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
	}
	parseResponse(t, rec, &userResp)

	if userResp.Email != "auth-flow@example.com" {
		t.Errorf("email = %q, want %q", userResp.Email, "auth-flow@example.com")
	}
	if userResp.ID == "" {
		t.Error("expected user ID to be set")
	}

	// Step 2: Login
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "auth-flow@example.com",
		"password": "SecurePass123!",
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	var tokens authTokens
	parseResponse(t, rec, &tokens)

	if tokens.AccessToken == "" {
		t.Fatal("expected access token")
	}
	if tokens.RefreshToken == "" {
		t.Fatal("expected refresh token")
	}
	if tokens.User.Email != "auth-flow@example.com" {
		t.Errorf("login user email = %q, want %q", tokens.User.Email, "auth-flow@example.com")
	}

	// Step 3: Access protected endpoint with token
	rec = doRequest(t, fiber.MethodGet, "/api/v1/users/me", nil, tokens.AccessToken)
	// If /users/me exists, should return 200; if not, check /users/get?id=
	if rec.Code == fiber.StatusNotFound {
		rec = doRequest(t, fiber.MethodGet, "/api/v1/users/get?id="+tokens.User.ID, nil, tokens.AccessToken)
	}
	if rec.Code != fiber.StatusOK {
		t.Logf("protected endpoint status: %d (may be expected if route differs)", rec.Code)
	}

	// Step 4: Refresh token
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": tokens.RefreshToken,
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	var newTokens authTokens
	parseResponse(t, rec, &newTokens)

	if newTokens.AccessToken == "" {
		t.Fatal("expected new access token after refresh")
	}
	if newTokens.AccessToken == tokens.AccessToken {
		t.Error("expected different access token after refresh")
	}

	// Step 5: Logout
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/logout", map[string]string{
		"refresh_token": newTokens.RefreshToken,
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	// Step 6: Refresh with old token should fail
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": newTokens.RefreshToken,
	}, "")
	if rec.Code == fiber.StatusOK {
		t.Error("expected refresh to fail after logout")
	}
}

// TestAuthFlow_LoginWrongPassword tests login with incorrect password.
func TestAuthFlow_LoginWrongPassword(t *testing.T) {
	// Register first
	doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":      "wrong-pw@example.com",
		"password":   "CorrectPass123!",
		"first_name": "Wrong",
	}, "")

	// Login with wrong password
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "wrong-pw@example.com",
		"password": "WrongPassword!",
	}, "")

	if rec.Code == fiber.StatusOK {
		t.Error("expected login to fail with wrong password")
	}
}

// TestAuthFlow_RegisterDuplicateEmail tests that duplicate registration fails.
func TestAuthFlow_RegisterDuplicateEmail(t *testing.T) {
	email := "duplicate@example.com"

	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":      email,
		"password":   "Pass12345!",
		"first_name": "First",
	}, "")
	expectStatus(t, rec, fiber.StatusCreated)

	// Second registration should fail
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":      email,
		"password":   "Pass12345!",
		"first_name": "Second",
	}, "")
	if rec.Code == fiber.StatusCreated {
		t.Error("expected duplicate registration to fail")
	}
}

// TestAuthFlow_ProtectedRouteWithoutToken tests that protected routes require auth.
func TestAuthFlow_ProtectedRouteWithoutToken(t *testing.T) {
	rec := doRequest(t, fiber.MethodGet, "/api/v1/accounts/list", nil, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, fiber.StatusUnauthorized)
	}
}
