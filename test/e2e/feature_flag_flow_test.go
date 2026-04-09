package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestFeatureFlagFlow_EvaluateWithAuth tests that authenticated users can evaluate feature flags.
func TestFeatureFlagFlow_EvaluateWithAuth(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("flag-eval"), "Password123!", "FlagEval")

	rec := doRequest(t, fiber.MethodPost, "/api/v1/flags/evaluate", map[string]string{
		"key": "some_feature_flag",
	}, user.AccessToken)
	// May return 404 if flag does not exist, but should not return 401 or 500
	if rec.Code == fiber.StatusUnauthorized {
		t.Error("authenticated user should not get 401")
	}
	if rec.Code == fiber.StatusInternalServerError {
		t.Error("expected graceful response, not 500")
	}
}

// TestFeatureFlagFlow_EvaluateWithoutAuth tests that flag evaluation requires authentication.
func TestFeatureFlagFlow_EvaluateWithoutAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodPost, "/api/v1/flags/evaluate", map[string]string{
		"key": "some_flag",
	}, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
