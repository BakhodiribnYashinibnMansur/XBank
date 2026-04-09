package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestCardFlow_IssueActivateBlock tests:
// issue card → activate (set PIN) → block → unblock
func TestCardFlow_IssueActivateBlock(t *testing.T) {
	user := registerAndLogin(t, "card-user@example.com", "Password123!", "CardUser")

	// Create account first
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/create", map[string]string{
		"currency": "UZS",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var acc struct{ ID string `json:"id"` }
	parseResponse(t, rec, &acc)

	// Step 1: Issue a debit card
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards", map[string]string{
		"account_id": acc.ID,
		"card_type":  "DEBIT",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var card struct {
		ID         string `json:"id"`
		AccountID  string `json:"account_id"`
		MaskedPAN  string `json:"masked_pan"`
		CardType   string `json:"card_type"`
		Status     string `json:"status"`
	}
	parseResponse(t, rec, &card)

	if card.ID == "" {
		t.Fatal("expected card ID")
	}
	if card.Status != "INACTIVE" {
		t.Errorf("initial status = %q, want INACTIVE", card.Status)
	}
	if card.CardType != "DEBIT" {
		t.Errorf("card_type = %q, want DEBIT", card.CardType)
	}
	if card.MaskedPAN == "" {
		t.Error("expected masked PAN")
	}

	// Step 2: Activate card (set PIN)
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+card.ID+"/activate", map[string]string{
		"pin": "1234",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	var activated struct {
		Status string `json:"status"`
	}
	parseResponse(t, rec, &activated)
	if activated.Status != "ACTIVE" {
		t.Errorf("activated status = %q, want ACTIVE", activated.Status)
	}

	// Step 3: Verify PIN
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+card.ID+"/verify-pin", map[string]string{
		"pin": "1234",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Step 4: Block card
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+card.ID+"/block", nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Step 5: Get card — should be BLOCKED
	rec = doHMACRequest(t, fiber.MethodGet, "/api/v1/cards/"+card.ID, nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	var blocked struct {
		Status string `json:"status"`
	}
	parseResponse(t, rec, &blocked)
	if blocked.Status != "BLOCKED" {
		t.Errorf("blocked status = %q, want BLOCKED", blocked.Status)
	}

	// Step 6: Unblock card
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+card.ID+"/unblock", nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestCardFlow_Tokenize tests card tokenization.
func TestCardFlow_Tokenize(t *testing.T) {
	user := registerAndLogin(t, "token-user@example.com", "Password123!", "TokenUser")

	// Create account + card
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/create", map[string]string{
		"currency": "UZS",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)
	var acc struct{ ID string `json:"id"` }
	parseResponse(t, rec, &acc)

	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards", map[string]string{
		"account_id": acc.ID,
		"card_type":  "VIRTUAL",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)
	var card struct{ ID string `json:"id"` }
	parseResponse(t, rec, &card)

	// Activate card first
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+card.ID+"/activate", map[string]string{
		"pin": "5678",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Tokenize
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+card.ID+"/tokenize", nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var tokenResp struct {
		Token    string `json:"token"`
		CardID   string `json:"card_id"`
		LastFour string `json:"last_four"`
		IsActive bool   `json:"is_active"`
	}
	parseResponse(t, rec, &tokenResp)

	if tokenResp.Token == "" {
		t.Error("expected token")
	}
	if tokenResp.CardID != card.ID {
		t.Errorf("card_id = %q, want %q", tokenResp.CardID, card.ID)
	}
	if !tokenResp.IsActive {
		t.Error("expected token to be active")
	}

	// List tokens
	rec = doHMACRequest(t, fiber.MethodGet, "/api/v1/cards/"+card.ID+"/tokens", nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestCardFlow_3DSEnroll tests 3D Secure enrollment.
func TestCardFlow_3DSEnroll(t *testing.T) {
	user := registerAndLogin(t, "threeds-user@example.com", "Password123!", "ThreeDS")

	// Create account + card + activate
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/create", map[string]string{
		"currency": "USD",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)
	var acc struct{ ID string `json:"id"` }
	parseResponse(t, rec, &acc)

	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards", map[string]string{
		"account_id": acc.ID,
		"card_type":  "DEBIT",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)
	var card struct{ ID string `json:"id"` }
	parseResponse(t, rec, &card)

	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+card.ID+"/activate", map[string]string{
		"pin": "9999",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Enroll 3DS
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+card.ID+"/3ds/enroll", map[string]string{
		"version": "2.2",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}
