package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestExchangeFlow_PublicRateQueries tests that exchange rate endpoints are publicly accessible.
func TestExchangeFlow_PublicRateQueries(t *testing.T) {
	t.Run("get_rate", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodGet, "/api/v1/currencies/rate?from=USD&to=UZS", nil, "")
		// May return 404 if no rates are seeded, but should not return 401 or 500
		if rec.Code == fiber.StatusUnauthorized {
			t.Error("exchange rate should be a public endpoint")
		}
		if rec.Code == fiber.StatusInternalServerError {
			t.Error("expected graceful response, not 500")
		}
	})

	t.Run("list_rates", func(t *testing.T) {
		rec := doRequest(t, fiber.MethodGet, "/api/v1/currencies/rates", nil, "")
		if rec.Code == fiber.StatusUnauthorized {
			t.Error("exchange rates listing should be a public endpoint")
		}
		if rec.Code == fiber.StatusInternalServerError {
			t.Error("expected graceful response, not 500")
		}
	})
}

// TestExchangeFlow_ConvertRequiresAuth tests that currency conversion requires authentication.
func TestExchangeFlow_ConvertRequiresAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodPost, "/api/v1/currencies/convert", map[string]interface{}{
		"from":   "USD",
		"to":     "UZS",
		"amount": 100,
	}, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated convert, got %d", rec.Code)
	}
}

// TestExchangeFlow_ConvertWithAuth tests currency conversion for authenticated users.
func TestExchangeFlow_ConvertWithAuth(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("convert-user"), "Password123!", "Convert")

	rec := doRequest(t, fiber.MethodPost, "/api/v1/currencies/convert", map[string]interface{}{
		"from":   "USD",
		"to":     "UZS",
		"amount": 100,
	}, user.AccessToken)
	// May return an error if no rates are configured, but should not be 401 or 500
	if rec.Code == fiber.StatusUnauthorized {
		t.Error("authenticated user should not get 401")
	}
	if rec.Code == fiber.StatusInternalServerError {
		t.Error("expected graceful error, not 500")
	}
}
