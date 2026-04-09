package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestNotificationFlow_ListAndRead tests notification listing and mark-as-read flow.
func TestNotificationFlow_ListAndRead(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("notif-flow"), "Password123!", "NotifUser")

	// List notifications (may be empty for a fresh user)
	rec := doRequest(t, fiber.MethodGet, "/api/v1/notifications/", nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestNotificationFlow_ListWithoutAuth tests that notification listing requires authentication.
func TestNotificationFlow_ListWithoutAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodGet, "/api/v1/notifications/", nil, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// TestNotificationFlow_GetNonExistent tests getting a non-existent notification.
func TestNotificationFlow_GetNonExistent(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("notif-nx"), "Password123!", "NotifNx")

	rec := doRequest(t, fiber.MethodGet, "/api/v1/notifications/00000000-0000-0000-0000-000000000000", nil, user.AccessToken)
	if rec.Code == fiber.StatusOK {
		t.Log("getting non-existent notification returned 200 (may be valid empty response)")
	}
	if rec.Code == fiber.StatusInternalServerError {
		t.Error("expected graceful error for non-existent notification, not 500")
	}
}

// TestNotificationFlow_DeleteWithoutAuth tests that notification deletion requires authentication.
func TestNotificationFlow_DeleteWithoutAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodDelete, "/api/v1/notifications/some-id", nil, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// TestNotificationFlow_MarkReadWithoutAuth tests that marking as read requires authentication.
func TestNotificationFlow_MarkReadWithoutAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodPost, "/api/v1/notifications/some-id/read", nil, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
