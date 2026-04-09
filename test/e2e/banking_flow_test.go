package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestBankingFlow_CreateAccountDepositTransferBalance tests:
// create account → deposit → transfer → check balance
func TestBankingFlow_CreateAccountDepositTransferBalance(t *testing.T) {
	// Setup: create two users
	user1 := registerAndLogin(t, "bank-user1@example.com", "Password123!", "Sender")
	user2 := registerAndLogin(t, "bank-user2@example.com", "Password123!", "Receiver")

	// Step 1: User1 creates an account
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/create", map[string]string{
		"currency": "UZS",
	}, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var account1 struct {
		ID             string `json:"id"`
		AccountNumber  string `json:"account_number"`
		Balance        int64  `json:"balance"`
		Currency       string `json:"currency"`
		Status         string `json:"status"`
	}
	parseResponse(t, rec, &account1)

	if account1.ID == "" {
		t.Fatal("expected account ID")
	}
	if account1.Balance != 0 {
		t.Errorf("initial balance = %d, want 0", account1.Balance)
	}
	if account1.Currency != "UZS" {
		t.Errorf("currency = %q, want UZS", account1.Currency)
	}
	if account1.Status != "ACTIVE" {
		t.Errorf("status = %q, want ACTIVE", account1.Status)
	}

	// Step 2: User2 creates an account
	rec = doRequest(t, fiber.MethodPost, "/api/v1/accounts/create", map[string]string{
		"currency": "UZS",
	}, user2.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var account2 struct {
		ID string `json:"id"`
	}
	parseResponse(t, rec, &account2)

	// Step 3: Deposit money to User1's account
	rec = doRequest(t, fiber.MethodPost, "/api/v1/accounts/deposit", map[string]interface{}{
		"account_id": account1.ID,
		"amount":     500_000, // 5000.00 UZS
	}, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	var depositResp struct {
		Balance int64 `json:"balance"`
	}
	parseResponse(t, rec, &depositResp)
	if depositResp.Balance != 500_000 {
		t.Errorf("balance after deposit = %d, want 500000", depositResp.Balance)
	}

	// Step 4: Transfer from User1 to User2
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": account1.ID,
		"to_account_id":   account2.ID,
		"amount":          100_000, // 1000.00 UZS
		"currency":        "UZS",
		"description":     "E2E test transfer",
	}, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var transferResp struct {
		ID            string `json:"id"`
		Status        string `json:"status"`
		Amount        int64  `json:"amount"`
		Description   string `json:"description"`
	}
	parseResponse(t, rec, &transferResp)

	if transferResp.ID == "" {
		t.Fatal("expected transfer ID")
	}
	if transferResp.Amount != 100_000 {
		t.Errorf("transfer amount = %d, want 100000", transferResp.Amount)
	}

	// Step 5: Check User1's account balance (should be 400,000)
	rec = doRequest(t, fiber.MethodGet, "/api/v1/accounts/get?id="+account1.ID, nil, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	var balance1 struct {
		Balance int64 `json:"balance"`
	}
	parseResponse(t, rec, &balance1)
	if balance1.Balance != 400_000 {
		t.Errorf("sender balance = %d, want 400000", balance1.Balance)
	}

	// Step 6: List user1's accounts
	rec = doRequest(t, fiber.MethodGet, "/api/v1/accounts/list", nil, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestBankingFlow_DepositAndWithdraw tests deposit and withdrawal.
func TestBankingFlow_DepositAndWithdraw(t *testing.T) {
	user := registerAndLogin(t, "deposit-withdraw@example.com", "Password123!", "DepWith")

	// Create account
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/create", map[string]string{
		"currency": "USD",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var acc struct{ ID string `json:"id"` }
	parseResponse(t, rec, &acc)

	// Deposit
	rec = doRequest(t, fiber.MethodPost, "/api/v1/accounts/deposit", map[string]interface{}{
		"account_id": acc.ID,
		"amount":     1_000_000,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Withdraw
	rec = doRequest(t, fiber.MethodPost, "/api/v1/accounts/withdraw", map[string]interface{}{
		"account_id": acc.ID,
		"amount":     300_000,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	var after struct{ Balance int64 `json:"balance"` }
	parseResponse(t, rec, &after)
	if after.Balance != 700_000 {
		t.Errorf("balance = %d, want 700000", after.Balance)
	}
}

// TestBankingFlow_InsufficientFunds tests transfer with insufficient balance.
func TestBankingFlow_InsufficientFunds(t *testing.T) {
	user1 := registerAndLogin(t, "poor-user@example.com", "Password123!", "Poor")
	user2 := registerAndLogin(t, "rich-user@example.com", "Password123!", "Rich")

	// Create accounts
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/create", map[string]string{"currency": "UZS"}, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)
	var acc1 struct{ ID string `json:"id"` }
	parseResponse(t, rec, &acc1)

	rec = doRequest(t, fiber.MethodPost, "/api/v1/accounts/create", map[string]string{"currency": "UZS"}, user2.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)
	var acc2 struct{ ID string `json:"id"` }
	parseResponse(t, rec, &acc2)

	// Try transfer without funds
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": acc1.ID,
		"to_account_id":   acc2.ID,
		"amount":          100,
		"currency":        "UZS",
	}, user1.AccessToken)

	if rec.Code == fiber.StatusCreated {
		t.Error("expected transfer to fail with insufficient funds")
	}
}
