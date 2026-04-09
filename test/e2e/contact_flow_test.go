package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestContactFlow_AddListDelete tests the full contact lifecycle:
// add contact -> list -> delete
func TestContactFlow_AddListDelete(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("contact-flow"), "Password123!", "ContactUser")

	// Add a contact
	rec := doRequest(t, fiber.MethodPost, "/api/v1/contacts/add", map[string]string{
		"name":  "Alice Contact",
		"phone": "+998901234567",
		"email": "alice@example.com",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var contact struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
	parseResponse(t, rec, &contact)

	if contact.ID == "" {
		t.Fatal("expected contact ID")
	}
	if contact.Name != "Alice Contact" {
		t.Errorf("name = %q, want %q", contact.Name, "Alice Contact")
	}

	// Add another contact
	rec = doRequest(t, fiber.MethodPost, "/api/v1/contacts/add", map[string]string{
		"name":  "Bob Contact",
		"phone": "+998907654321",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	// List contacts
	rec = doRequest(t, fiber.MethodGet, "/api/v1/contacts/list", nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Delete first contact
	rec = doRequest(t, fiber.MethodDelete, "/api/v1/contacts/delete?id="+contact.ID, nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestContactFlow_AddWithoutAuth tests that contact creation requires authentication.
func TestContactFlow_AddWithoutAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodPost, "/api/v1/contacts/add", map[string]string{
		"name":  "Unauth",
		"phone": "+998900000000",
	}, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// TestContactFlow_ListWithoutAuth tests that contact listing requires authentication.
func TestContactFlow_ListWithoutAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodGet, "/api/v1/contacts/list", nil, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// TestContactFlow_CrossUserIsolation tests that users cannot see each other's contacts.
func TestContactFlow_CrossUserIsolation(t *testing.T) {
	user1 := registerAndLogin(t, uniqueEmail("con-iso1"), "Password123!", "ConIso1")
	user2 := registerAndLogin(t, uniqueEmail("con-iso2"), "Password123!", "ConIso2")

	// User1 adds a contact
	rec := doRequest(t, fiber.MethodPost, "/api/v1/contacts/add", map[string]string{
		"name":  "Private Contact",
		"phone": "+998901112233",
	}, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	// User2 lists contacts -- should not see user1's contacts
	rec = doRequest(t, fiber.MethodGet, "/api/v1/contacts/list", nil, user2.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}
