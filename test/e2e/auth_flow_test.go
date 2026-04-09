package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestAuthFlow_RegisterLoginRefreshLogout tests the complete auth lifecycle:
// register -> login -> refresh -> logout
func TestAuthFlow_RegisterLoginRefreshLogout(t *testing.T) {
	email := uniqueEmail("auth-flow")

	// Step 1: Register
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":      email,
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

	if userResp.Email != email {
		t.Errorf("email = %q, want %q", userResp.Email, email)
	}
	if userResp.ID == "" {
		t.Error("expected user ID to be set")
	}

	// Step 2: Login
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
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
	if tokens.User.Email != email {
		t.Errorf("login user email = %q, want %q", tokens.User.Email, email)
	}

	// Step 3: Access protected endpoint with token
	rec = doRequest(t, fiber.MethodGet, "/api/v1/users/get?id="+tokens.User.ID, nil, tokens.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

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

	// Step 5: Use new access token on protected endpoint
	rec = doRequest(t, fiber.MethodGet, "/api/v1/users/get?id="+tokens.User.ID, nil, newTokens.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Step 6: Logout
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/logout", map[string]string{
		"refresh_token": newTokens.RefreshToken,
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	// Step 7: Refresh with old token should fail
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": newTokens.RefreshToken,
	}, "")
	if rec.Code == fiber.StatusOK {
		t.Error("expected refresh to fail after logout")
	}
}

// TestAuthFlow_LoginWrongPassword tests login with incorrect password.
func TestAuthFlow_LoginWrongPassword(t *testing.T) {
	email := uniqueEmail("wrong-pw")
	doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":      email,
		"password":   "CorrectPass123!",
		"first_name": "Wrong",
		"last_name":  "Test",
	}, "")

	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": "WrongPassword!",
	}, "")

	if rec.Code == fiber.StatusOK {
		t.Error("expected login to fail with wrong password")
	}
}

// TestAuthFlow_LoginNonExistentUser tests login with an email that does not exist.
func TestAuthFlow_LoginNonExistentUser(t *testing.T) {
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "nonexistent-user-12345@example.com",
		"password": "AnyPassword123!",
	}, "")
	if rec.Code == fiber.StatusOK {
		t.Error("expected login with non-existent email to fail")
	}
}

// TestAuthFlow_LoginEmptyCredentials tests login with empty credentials.
func TestAuthFlow_LoginEmptyCredentials(t *testing.T) {
	t.Run("empty_email", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
			"email":    "",
			"password": "Password123!",
		}, "")
		if rec.Code == fiber.StatusOK {
			t.Error("expected login with empty email to fail")
		}
	})

	t.Run("empty_password", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
			"email":    "someone@example.com",
			"password": "",
		}, "")
		if rec.Code == fiber.StatusOK {
			t.Error("expected login with empty password to fail")
		}
	})

	t.Run("empty_body", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{}, "")
		if rec.Code == fiber.StatusOK {
			t.Error("expected login with empty body to fail")
		}
	})
}

// TestAuthFlow_RegisterDuplicateEmail tests that duplicate registration fails.
func TestAuthFlow_RegisterDuplicateEmail(t *testing.T) {
	email := uniqueEmail("duplicate")

	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":      email,
		"password":   "Pass12345!",
		"first_name": "First",
		"last_name":  "Test",
	}, "")
	expectStatus(t, rec, fiber.StatusCreated)

	// Second registration should fail
	rec = doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":      email,
		"password":   "Pass12345!",
		"first_name": "Second",
		"last_name":  "Test",
	}, "")
	if rec.Code == fiber.StatusCreated {
		t.Error("expected duplicate registration to fail")
	}
}

// TestAuthFlow_ProtectedRouteWithoutToken tests that protected routes require auth.
func TestAuthFlow_ProtectedRouteWithoutToken(t *testing.T) {
	protectedRoutes := []struct {
		method string
		path   string
	}{
		{fiber.MethodGet, "/api/v1/accounts/list"},
		{fiber.MethodGet, "/api/v1/users/get?id=some-id"},
		{fiber.MethodPost, "/api/v1/accounts/create"},
		{fiber.MethodGet, "/api/v1/kyc/status"},
		{fiber.MethodGet, "/api/v1/notifications/"},
		{fiber.MethodGet, "/api/v1/beneficiaries/list"},
		{fiber.MethodGet, "/api/v1/contacts/list"},
	}

	for _, rt := range protectedRoutes {
		t.Run(rt.method+"_"+rt.path, func(t *testing.T) {
			rec := doRequest(t, rt.method, rt.path, nil, "")
			if rec.Code != fiber.StatusUnauthorized {
				t.Errorf("status = %d, want %d for %s %s", rec.Code, fiber.StatusUnauthorized, rt.method, rt.path)
			}
		})
	}
}

// TestAuthFlow_ProtectedRouteWithMalformedToken tests various malformed tokens.
func TestAuthFlow_ProtectedRouteWithMalformedToken(t *testing.T) {
	invalidTokens := []struct {
		name  string
		token string
	}{
		{"random_string", "not-a-jwt-token"},
		{"empty", ""},
		{"partial_jwt", "eyJhbGciOiJFUzI1NiJ9.incomplete"},
		{"three_dots", "a.b.c"},
	}

	for _, tt := range invalidTokens {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, fiber.MethodGet, "/api/v1/accounts/list", nil, tt.token)
			if rec.Code != fiber.StatusUnauthorized {
				t.Errorf("status = %d, want %d for token %q", rec.Code, fiber.StatusUnauthorized, tt.name)
			}
		})
	}
}

// TestAuthFlow_MultipleLoginsReturnDifferentTokens tests that each login returns unique tokens.
func TestAuthFlow_MultipleLoginsReturnDifferentTokens(t *testing.T) {
	email := uniqueEmail("multi-login")
	password := "Password123!"

	doRequest(t, fiber.MethodPost, "/api/v1/auth/register", map[string]string{
		"email":      email,
		"password":   password,
		"first_name": "Multi",
		"last_name":  "Login",
	}, "")

	// Login twice
	rec1 := doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": email, "password": password,
	}, "")
	expectStatus(t, rec1, fiber.StatusOK)
	var tokens1 authTokens
	parseResponse(t, rec1, &tokens1)

	rec2 := doRequest(t, fiber.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": email, "password": password,
	}, "")
	expectStatus(t, rec2, fiber.StatusOK)
	var tokens2 authTokens
	parseResponse(t, rec2, &tokens2)

	if tokens1.AccessToken == tokens2.AccessToken {
		t.Error("expected different access tokens for separate logins")
	}
	if tokens1.RefreshToken == tokens2.RefreshToken {
		t.Error("expected different refresh tokens for separate logins")
	}
}

// TestAuthFlow_RefreshDoesNotRequireAccessToken tests that refresh endpoint is public.
func TestAuthFlow_RefreshDoesNotRequireAccessToken(t *testing.T) {
	email := uniqueEmail("refresh-public")
	tokens := registerAndLogin(t, email, "Password123!", "RefreshPub")

	// Refresh without providing an Authorization header
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/refresh", map[string]string{
		"refresh_token": tokens.RefreshToken,
	}, "")
	expectStatus(t, rec, fiber.StatusOK)

	var newTokens authTokens
	parseResponse(t, rec, &newTokens)
	if newTokens.AccessToken == "" {
		t.Fatal("expected access token from refresh")
	}
}
