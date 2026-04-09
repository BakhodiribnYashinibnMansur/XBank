package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestCardFlow_IssueActivateBlock tests:
// issue card -> activate (set PIN) -> block -> unblock
func TestCardFlow_IssueActivateBlock(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("card-user"), "Password123!", "CardUser")

	// Create account first
	accID := createAccountWithDeposit(t, user.AccessToken, "UZS", 0)

	// Step 1: Issue a debit card
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/cards", map[string]string{
		"account_id": accID,
		"card_type":  "DEBIT",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var card struct {
		ID        string `json:"id"`
		AccountID string `json:"account_id"`
		MaskedPAN string `json:"masked_pan"`
		CardType  string `json:"card_type"`
		Status    string `json:"status"`
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

	// Step 5: Get card -- should be BLOCKED
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

	// Step 7: Get card -- should be ACTIVE again
	rec = doHMACRequest(t, fiber.MethodGet, "/api/v1/cards/"+card.ID, nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	var unblocked struct {
		Status string `json:"status"`
	}
	parseResponse(t, rec, &unblocked)
	if unblocked.Status != "ACTIVE" {
		t.Errorf("unblocked status = %q, want ACTIVE", unblocked.Status)
	}
}

// TestCardFlow_VerifyWrongPIN tests that verifying with a wrong PIN fails.
func TestCardFlow_VerifyWrongPIN(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("wrong-pin"), "Password123!", "WrongPIN")
	cardID, _ := issueAndActivateCard(t, user.AccessToken, "UZS", "DEBIT", "1234")

	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+cardID+"/verify-pin", map[string]string{
		"pin": "9999",
	}, user.AccessToken)

	if rec.Code == fiber.StatusOK {
		t.Error("expected wrong PIN verification to fail")
	}
}

// TestCardFlow_ChangePIN tests changing the card PIN.
func TestCardFlow_ChangePIN(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("change-pin"), "Password123!", "ChgPIN")
	cardID, _ := issueAndActivateCard(t, user.AccessToken, "UZS", "DEBIT", "1111")

	// Change PIN
	rec := doHMACRequest(t, fiber.MethodPut, "/api/v1/cards/"+cardID+"/pin", map[string]string{
		"old_pin": "1111",
		"new_pin": "2222",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Verify with old PIN should fail
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+cardID+"/verify-pin", map[string]string{
		"pin": "1111",
	}, user.AccessToken)
	if rec.Code == fiber.StatusOK {
		t.Error("expected old PIN verification to fail after PIN change")
	}

	// Verify with new PIN should succeed
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+cardID+"/verify-pin", map[string]string{
		"pin": "2222",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestCardFlow_Tokenize tests card tokenization.
func TestCardFlow_Tokenize(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("token-user"), "Password123!", "TokenUser")
	cardID, _ := issueAndActivateCard(t, user.AccessToken, "UZS", "VIRTUAL", "5678")

	// Tokenize
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+cardID+"/tokenize", nil, user.AccessToken)
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
	if tokenResp.CardID != cardID {
		t.Errorf("card_id = %q, want %q", tokenResp.CardID, cardID)
	}
	if !tokenResp.IsActive {
		t.Error("expected token to be active")
	}

	// List tokens
	rec = doHMACRequest(t, fiber.MethodGet, "/api/v1/cards/"+cardID+"/tokens", nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Revoke token
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/tokens/revoke", map[string]string{
		"token": tokenResp.Token,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestCardFlow_3DSEnroll tests 3D Secure enrollment.
func TestCardFlow_3DSEnroll(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("threeds-user"), "Password123!", "ThreeDS")
	cardID, _ := issueAndActivateCard(t, user.AccessToken, "USD", "DEBIT", "9999")

	// Enroll 3DS
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+cardID+"/3ds/enroll", map[string]string{
		"version": "2.2",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestCardFlow_ListByAccount tests listing all cards for an account.
func TestCardFlow_ListByAccount(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("list-cards"), "Password123!", "ListCards")
	accID := createAccountWithDeposit(t, user.AccessToken, "UZS", 0)

	// Issue two cards for the same account
	cardTypes := []string{"DEBIT", "VIRTUAL"}
	for _, ct := range cardTypes {
		rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/cards", map[string]string{
			"account_id": accID,
			"card_type":  ct,
		}, user.AccessToken)
		expectStatus(t, rec, fiber.StatusCreated)
	}

	// List cards by account
	rec := doHMACRequest(t, fiber.MethodGet, "/api/v1/cards/?account_id="+accID, nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestCardFlow_BlockAlreadyBlocked tests blocking a card that is already blocked.
func TestCardFlow_BlockAlreadyBlocked(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("double-block"), "Password123!", "DblBlock")
	cardID, _ := issueAndActivateCard(t, user.AccessToken, "UZS", "DEBIT", "1234")

	// Block
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+cardID+"/block", nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Block again -- should either succeed (idempotent) or return specific error
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+cardID+"/block", nil, user.AccessToken)
	if rec.Code == fiber.StatusInternalServerError {
		t.Error("expected graceful handling of double block, not 500")
	}
}

// TestCardFlow_UnblockActiveCard tests unblocking a card that is not blocked.
func TestCardFlow_UnblockActiveCard(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("unblock-active"), "Password123!", "UnblkAct")
	cardID, _ := issueAndActivateCard(t, user.AccessToken, "UZS", "DEBIT", "1234")

	// Unblock active card -- should either succeed (idempotent) or return specific error
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+cardID+"/unblock", nil, user.AccessToken)
	if rec.Code == fiber.StatusInternalServerError {
		t.Error("expected graceful handling of unblock on active card, not 500")
	}
}

// TestCardFlow_IssueCardWithoutAuth tests that card issuance requires authentication.
func TestCardFlow_IssueCardWithoutAuth(t *testing.T) {
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/cards", map[string]string{
		"account_id": "some-id",
		"card_type":  "DEBIT",
	}, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// TestCardFlow_Hold tests card hold create, capture, and release.
func TestCardFlow_Hold(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("card-hold"), "Password123!", "Hold")
	cardID, accountID := issueAndActivateCard(t, user.AccessToken, "UZS", "DEBIT", "4444")

	// Deposit funds to the account
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/deposit", map[string]interface{}{
		"account_id": accountID,
		"amount":     500_000,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Create a hold
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/holds", map[string]interface{}{
		"card_id": cardID,
		"amount":  100_000,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var hold struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Amount int64  `json:"amount"`
	}
	parseResponse(t, rec, &hold)

	if hold.ID == "" {
		t.Fatal("expected hold ID")
	}

	// List holds for the card
	rec = doHMACRequest(t, fiber.MethodGet, "/api/v1/cards/"+cardID+"/holds", nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Release the hold
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/holds/"+hold.ID+"/release", nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestCardFlow_ChangePINWrongOld tests changing PIN with wrong old PIN.
func TestCardFlow_ChangePINWrongOld(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("chgpin-wrong"), "Password123!", "ChgWrong")
	cardID, _ := issueAndActivateCard(t, user.AccessToken, "UZS", "DEBIT", "1234")

	rec := doHMACRequest(t, fiber.MethodPut, "/api/v1/cards/"+cardID+"/pin", map[string]string{
		"old_pin": "9999",
		"new_pin": "5555",
	}, user.AccessToken)

	if rec.Code == fiber.StatusOK {
		t.Error("expected PIN change with wrong old PIN to fail")
	}
}

// TestCardFlow_EMVSet tests setting EMV data on a card.
func TestCardFlow_EMVSet(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("emv-user"), "Password123!", "EMV")
	cardID, _ := issueAndActivateCard(t, user.AccessToken, "UZS", "DEBIT", "3333")

	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+cardID+"/emv", map[string]string{
		"application_id": "A0000000041010",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestCardFlow_AnotherUserCannotAccessCard tests cross-user card access is denied.
func TestCardFlow_AnotherUserCannotAccessCard(t *testing.T) {
	owner := registerAndLogin(t, uniqueEmail("card-owner"), "Password123!", "Owner")
	attacker := registerAndLogin(t, uniqueEmail("card-attacker"), "Password123!", "Attacker")

	cardID, _ := issueAndActivateCard(t, owner.AccessToken, "UZS", "DEBIT", "1234")

	// Attacker tries to block owner's card
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/cards/"+cardID+"/block", nil, attacker.AccessToken)
	if rec.Code == fiber.StatusOK {
		t.Error("expected attacker to be denied blocking another user's card")
	}

	// Attacker tries to get owner's card details
	rec = doHMACRequest(t, fiber.MethodGet, "/api/v1/cards/"+cardID, nil, attacker.AccessToken)
	if rec.Code == fiber.StatusOK {
		t.Error("expected attacker to be denied viewing another user's card")
	}
}
