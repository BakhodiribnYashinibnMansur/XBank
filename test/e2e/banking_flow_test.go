package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestBankingFlow_CreateAccountDepositTransferBalance tests:
// create account -> deposit -> transfer -> check balance
func TestBankingFlow_CreateAccountDepositTransferBalance(t *testing.T) {
	// Setup: create two users
	user1 := registerAndLogin(t, uniqueEmail("bank-user1"), "Password123!", "Sender")
	user2 := registerAndLogin(t, uniqueEmail("bank-user2"), "Password123!", "Receiver")

	// Step 1: User1 creates an account
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/create", map[string]string{
		"currency": "UZS",
	}, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var account1 struct {
		ID            string `json:"id"`
		AccountNumber string `json:"account_number"`
		Balance       int64  `json:"balance"`
		Currency      string `json:"currency"`
		Status        string `json:"status"`
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
	acc2ID := createAccountWithDeposit(t, user2.AccessToken, "UZS", 0)

	// Step 3: Deposit money to User1's account
	rec = doRequest(t, fiber.MethodPost, "/api/v1/accounts/deposit", map[string]interface{}{
		"account_id": account1.ID,
		"amount":     500_000,
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
		"to_account_id":   acc2ID,
		"amount":          100_000,
		"currency":        "UZS",
		"description":     "E2E test transfer",
	}, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var transferResp struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		Amount      int64  `json:"amount"`
		Description string `json:"description"`
	}
	parseResponse(t, rec, &transferResp)

	if transferResp.ID == "" {
		t.Fatal("expected transfer ID")
	}
	if transferResp.Amount != 100_000 {
		t.Errorf("transfer amount = %d, want 100000", transferResp.Amount)
	}

	// Step 5: Check User1's account balance (should be 400,000)
	balance := getAccountBalance(t, user1.AccessToken, account1.ID)
	if balance != 400_000 {
		t.Errorf("sender balance = %d, want 400000", balance)
	}

	// Step 6: List user1's accounts
	rec = doRequest(t, fiber.MethodGet, "/api/v1/accounts/list", nil, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestBankingFlow_DepositAndWithdraw tests deposit and withdrawal.
func TestBankingFlow_DepositAndWithdraw(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("deposit-withdraw"), "Password123!", "DepWith")
	accID := createAccountWithDeposit(t, user.AccessToken, "USD", 1_000_000)

	// Withdraw
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/withdraw", map[string]interface{}{
		"account_id": accID,
		"amount":     300_000,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	var after struct {
		Balance int64 `json:"balance"`
	}
	parseResponse(t, rec, &after)
	if after.Balance != 700_000 {
		t.Errorf("balance = %d, want 700000", after.Balance)
	}
}

// TestBankingFlow_InsufficientFunds tests transfer with insufficient balance.
func TestBankingFlow_InsufficientFunds(t *testing.T) {
	user1 := registerAndLogin(t, uniqueEmail("poor-user"), "Password123!", "Poor")
	user2 := registerAndLogin(t, uniqueEmail("rich-user"), "Password123!", "Rich")

	acc1 := createAccountWithDeposit(t, user1.AccessToken, "UZS", 0)
	acc2 := createAccountWithDeposit(t, user2.AccessToken, "UZS", 0)

	// Try transfer without funds
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": acc1,
		"to_account_id":   acc2,
		"amount":          100,
		"currency":        "UZS",
	}, user1.AccessToken)

	if rec.Code == fiber.StatusCreated {
		t.Error("expected transfer to fail with insufficient funds")
	}
}

// TestBankingFlow_FullLifecycle tests a complete banking lifecycle:
// register -> create account -> deposit -> issue card -> activate -> transfer -> check history
func TestBankingFlow_FullLifecycle(t *testing.T) {
	user1 := registerAndLogin(t, uniqueEmail("lifecycle1"), "Password123!", "Life1")
	user2 := registerAndLogin(t, uniqueEmail("lifecycle2"), "Password123!", "Life2")

	// Create accounts
	acc1 := createAccountWithDeposit(t, user1.AccessToken, "UZS", 2_000_000)
	acc2 := createAccountWithDeposit(t, user2.AccessToken, "UZS", 0)

	// Issue and activate a card for user1
	cardID, _ := issueAndActivateCard(t, user1.AccessToken, "UZS", "DEBIT", "1234")
	_ = cardID

	// Make multiple transfers
	for i := 0; i < 3; i++ {
		rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
			"from_account_id": acc1,
			"to_account_id":   acc2,
			"amount":          100_000,
			"currency":        "UZS",
			"description":     "lifecycle transfer",
		}, user1.AccessToken)
		expectStatus(t, rec, fiber.StatusCreated)
	}

	// Check final balance: 2M - 3 * 100k = 1.7M
	bal := getAccountBalance(t, user1.AccessToken, acc1)
	if bal != 1_700_000 {
		t.Errorf("final balance = %d, want 1700000", bal)
	}

	// Check history
	rec := doRequest(t, fiber.MethodGet, "/api/v1/accounts/history?id="+acc1, nil, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Check transfer history
	rec = doHMACRequest(t, fiber.MethodGet, "/api/v1/transfers/list?account_id="+acc1, nil, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}
