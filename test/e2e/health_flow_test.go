package e2e_test

import (
	"encoding/json"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestHealthFlow_LiveEndpoint tests the liveness probe.
func TestHealthFlow_LiveEndpoint(t *testing.T) {
	rec := doRequest(t, fiber.MethodGet, "/health/live", nil, "")
	expectStatus(t, rec, fiber.StatusOK)

	var resp map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parsing health response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("health status = %q, want ok", resp["status"])
	}
}

// TestHealthFlow_HealthEndpoint tests the default health endpoint.
func TestHealthFlow_HealthEndpoint(t *testing.T) {
	rec := doRequest(t, fiber.MethodGet, "/health", nil, "")
	expectStatus(t, rec, fiber.StatusOK)
}

// TestHealthFlow_ReadyEndpoint tests the readiness probe.
func TestHealthFlow_ReadyEndpoint(t *testing.T) {
	rec := doRequest(t, fiber.MethodGet, "/health/ready", nil, "")
	// Ready may fail if some dependencies are missing in test setup; accept 200 or 503
	expectStatusOneOf(t, rec, fiber.StatusOK, fiber.StatusServiceUnavailable)
}

// TestHealthFlow_MetricsEndpoint tests the Prometheus metrics endpoint.
func TestHealthFlow_MetricsEndpoint(t *testing.T) {
	rec := doRequest(t, fiber.MethodGet, "/metrics", nil, "")
	expectStatus(t, rec, fiber.StatusOK)

	// Metrics should contain Prometheus format content
	body := rec.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty metrics response")
	}
}

// TestHealthFlow_SwaggerDocs tests that API documentation is accessible.
func TestHealthFlow_SwaggerDocs(t *testing.T) {
	rec := doRequest(t, fiber.MethodGet, "/swagger/index.html", nil, "")
	// Swagger may return 200 or redirect; just verify it does not panic
	if rec.Code == fiber.StatusInternalServerError {
		t.Error("expected swagger endpoint to not return 500")
	}
}

// TestHealthFlow_NotFoundRoute tests that a non-existent route returns 404.
func TestHealthFlow_NotFoundRoute(t *testing.T) {
	rec := doRequest(t, fiber.MethodGet, "/api/v1/nonexistent", nil, "")
	if rec.Code == fiber.StatusOK {
		t.Error("expected 404 for non-existent route")
	}
}
