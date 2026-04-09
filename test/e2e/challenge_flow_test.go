package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestChallengeFlow_RequestWithoutAuth tests that challenge request requires authentication.
func TestChallengeFlow_RequestWithoutAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/challenge/request", map[string]string{
		"action": "transfer",
	}, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated challenge request, got %d", rec.Code)
	}
}

// TestChallengeFlow_VerifyWithoutAuth tests that challenge verify requires authentication.
func TestChallengeFlow_VerifyWithoutAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/challenge/verify", map[string]string{
		"token": "some-token",
	}, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated challenge verify, got %d", rec.Code)
	}
}

// TestChallengeFlow_RequestWithAuth tests requesting a step-up auth challenge.
func TestChallengeFlow_RequestWithAuth(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("challenge"), "Password123!", "Challenge")

	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/challenge/request", map[string]string{
		"action": "transfer",
	}, user.AccessToken)
	// The challenge endpoint should return a challenge token or similar
	expectStatusOneOf(t, rec, fiber.StatusOK, fiber.StatusCreated)

	var challengeResp struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expires_at"`
	}
	parseResponse(t, rec, &challengeResp)

	if challengeResp.Token == "" {
		t.Error("expected challenge token")
	}
}

// TestChallengeFlow_VerifyWithInvalidToken tests verifying an invalid challenge token.
func TestChallengeFlow_VerifyWithInvalidToken(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("chal-invalid"), "Password123!", "ChalInv")

	rec := doRequest(t, fiber.MethodPost, "/api/v1/auth/challenge/verify", map[string]string{
		"token": "invalid-challenge-token",
	}, user.AccessToken)

	if rec.Code == fiber.StatusOK {
		t.Error("expected invalid challenge token verification to fail")
	}
}
