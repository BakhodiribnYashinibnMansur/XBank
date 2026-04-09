package e2e_test

import (
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TestTransferFlow_SendAndVerifyBalances tests end-to-end money transfer
// with balance verification on both sender and receiver sides.
func TestTransferFlow_SendAndVerifyBalances(t *testing.T) {
	sender := registerAndLogin(t, uniqueEmail("xfer-sender"), "Password123!", "Sender")
	receiver := registerAndLogin(t, uniqueEmail("xfer-receiver"), "Password123!", "Receiver")

	senderAccID := createAccountWithDeposit(t, sender.AccessToken, "UZS", 1_000_000)
	receiverAccID := createAccountWithDeposit(t, receiver.AccessToken, "UZS", 0)

	// Transfer
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": senderAccID,
		"to_account_id":   receiverAccID,
		"amount":          250_000,
		"currency":        "UZS",
		"description":     "Transfer flow test",
	}, sender.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var transfer struct {
		ID          string `json:"id"`
		Status      string `json:"status"`
		Amount      int64  `json:"amount"`
		Description string `json:"description"`
	}
	parseResponse(t, rec, &transfer)

	if transfer.ID == "" {
		t.Fatal("expected transfer ID")
	}
	if transfer.Amount != 250_000 {
		t.Errorf("transfer amount = %d, want 250000", transfer.Amount)
	}

	// Check sender balance (should be 750,000)
	senderBal := getAccountBalance(t, sender.AccessToken, senderAccID)
	if senderBal != 750_000 {
		t.Errorf("sender balance = %d, want 750000", senderBal)
	}

	// Check receiver balance (should be 250,000)
	receiverBal := getAccountBalance(t, receiver.AccessToken, receiverAccID)
	if receiverBal != 250_000 {
		t.Errorf("receiver balance = %d, want 250000", receiverBal)
	}

	// Get transfer by ID
	rec = doHMACRequest(t, fiber.MethodGet, "/api/v1/transfers/get?id="+transfer.ID, nil, sender.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
	var fetchedTransfer struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	parseResponse(t, rec, &fetchedTransfer)
	if fetchedTransfer.ID != transfer.ID {
		t.Errorf("fetched transfer ID = %q, want %q", fetchedTransfer.ID, transfer.ID)
	}
}

// TestTransferFlow_MultipleTransfers tests multiple sequential transfers and balance consistency.
func TestTransferFlow_MultipleTransfers(t *testing.T) {
	user1 := registerAndLogin(t, uniqueEmail("multi-xfer1"), "Password123!", "Multi1")
	user2 := registerAndLogin(t, uniqueEmail("multi-xfer2"), "Password123!", "Multi2")

	acc1 := createAccountWithDeposit(t, user1.AccessToken, "UZS", 500_000)
	acc2 := createAccountWithDeposit(t, user2.AccessToken, "UZS", 500_000)

	// Transfer 100k from user1 -> user2
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": acc1, "to_account_id": acc2,
		"amount": 100_000, "currency": "UZS", "description": "first",
	}, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	// Transfer 50k from user2 -> user1
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": acc2, "to_account_id": acc1,
		"amount": 50_000, "currency": "UZS", "description": "second",
	}, user2.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	// Transfer 75k from user1 -> user2
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": acc1, "to_account_id": acc2,
		"amount": 75_000, "currency": "UZS", "description": "third",
	}, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	// Final balances:
	// user1: 500k - 100k + 50k - 75k = 375k
	// user2: 500k + 100k - 50k + 75k = 625k
	bal1 := getAccountBalance(t, user1.AccessToken, acc1)
	if bal1 != 375_000 {
		t.Errorf("user1 balance = %d, want 375000", bal1)
	}

	bal2 := getAccountBalance(t, user2.AccessToken, acc2)
	if bal2 != 625_000 {
		t.Errorf("user2 balance = %d, want 625000", bal2)
	}
}

// TestTransferFlow_TransferHistory tests listing transfers for an account.
func TestTransferFlow_TransferHistory(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("xfer-hist"), "Password123!", "Hist")
	user2 := registerAndLogin(t, uniqueEmail("xfer-hist2"), "Password123!", "Hist2")

	acc := createAccountWithDeposit(t, user.AccessToken, "UZS", 1_000_000)
	acc2 := createAccountWithDeposit(t, user2.AccessToken, "UZS", 0)

	// Make a few transfers
	for i := 0; i < 3; i++ {
		rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
			"from_account_id": acc, "to_account_id": acc2,
			"amount": 10_000, "currency": "UZS", "description": "history test",
		}, user.AccessToken)
		expectStatus(t, rec, fiber.StatusCreated)
	}

	// List transfers
	rec := doHMACRequest(t, fiber.MethodGet, "/api/v1/transfers/list?account_id="+acc, nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// History endpoint
	rec = doHMACRequest(t, fiber.MethodGet, "/api/v1/transfers/history?account_id="+acc, nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestTransferFlow_ScheduledTransfer tests scheduling a transfer for the future.
func TestTransferFlow_ScheduledTransfer(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("sched-xfer"), "Password123!", "Sched")
	user2 := registerAndLogin(t, uniqueEmail("sched-xfer2"), "Password123!", "Sched2")

	acc1 := createAccountWithDeposit(t, user.AccessToken, "UZS", 500_000)
	acc2 := createAccountWithDeposit(t, user2.AccessToken, "UZS", 0)

	// Schedule a transfer for tomorrow
	executeAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/scheduled/", map[string]interface{}{
		"from_account_id": acc1,
		"to_account_id":   acc2,
		"amount":          100_000,
		"currency":        "UZS",
		"description":     "scheduled test",
		"execute_at":      executeAt,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var scheduled struct {
		ID        string `json:"id"`
		Status    string `json:"status"`
		ExecuteAt string `json:"execute_at"`
	}
	parseResponse(t, rec, &scheduled)

	if scheduled.ID == "" {
		t.Fatal("expected scheduled transfer ID")
	}

	// List scheduled transfers
	rec = doHMACRequest(t, fiber.MethodGet, "/api/v1/transfers/scheduled/list?account_id="+acc1, nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Get scheduled transfer by ID
	rec = doHMACRequest(t, fiber.MethodGet, "/api/v1/transfers/scheduled/get?id="+scheduled.ID, nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Cancel scheduled transfer
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/scheduled/cancel", map[string]interface{}{
		"id": scheduled.ID,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Verify balance unchanged (transfer was cancelled, not executed)
	balance := getAccountBalance(t, user.AccessToken, acc1)
	if balance != 500_000 {
		t.Errorf("balance after cancelled scheduled transfer = %d, want 500000", balance)
	}
}

// TestTransferFlow_SelfTransferFails tests that transferring to the same account is rejected.
func TestTransferFlow_SelfTransferFails(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("self-xfer"), "Password123!", "Self")
	accID := createAccountWithDeposit(t, user.AccessToken, "UZS", 100_000)

	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": accID,
		"to_account_id":   accID,
		"amount":          10_000,
		"currency":        "UZS",
	}, user.AccessToken)

	if rec.Code == fiber.StatusCreated {
		t.Error("expected self-transfer to be rejected")
	}
}

// TestTransferFlow_ZeroAmountFails tests that zero-amount transfer is rejected.
func TestTransferFlow_ZeroAmountFails(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("zero-xfer"), "Password123!", "Zero")
	user2 := registerAndLogin(t, uniqueEmail("zero-xfer2"), "Password123!", "Zero2")

	acc1 := createAccountWithDeposit(t, user.AccessToken, "UZS", 100_000)
	acc2 := createAccountWithDeposit(t, user2.AccessToken, "UZS", 0)

	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": acc1,
		"to_account_id":   acc2,
		"amount":          0,
		"currency":        "UZS",
	}, user.AccessToken)

	if rec.Code == fiber.StatusCreated {
		t.Error("expected zero-amount transfer to be rejected")
	}
}

// TestTransferFlow_NegativeAmountFails tests that negative-amount transfer is rejected.
func TestTransferFlow_NegativeAmountFails(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("neg-xfer"), "Password123!", "Neg")
	user2 := registerAndLogin(t, uniqueEmail("neg-xfer2"), "Password123!", "Neg2")

	acc1 := createAccountWithDeposit(t, user.AccessToken, "UZS", 100_000)
	acc2 := createAccountWithDeposit(t, user2.AccessToken, "UZS", 0)

	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": acc1,
		"to_account_id":   acc2,
		"amount":          -50_000,
		"currency":        "UZS",
	}, user.AccessToken)

	if rec.Code == fiber.StatusCreated {
		t.Error("expected negative-amount transfer to be rejected")
	}
}

// TestTransferFlow_InsufficientFunds tests transfer with insufficient balance.
func TestTransferFlow_InsufficientFunds(t *testing.T) {
	user1 := registerAndLogin(t, uniqueEmail("insuf-xfer"), "Password123!", "Insuf")
	user2 := registerAndLogin(t, uniqueEmail("insuf-xfer2"), "Password123!", "Insuf2")

	acc1 := createAccountWithDeposit(t, user1.AccessToken, "UZS", 10_000)
	acc2 := createAccountWithDeposit(t, user2.AccessToken, "UZS", 0)

	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": acc1,
		"to_account_id":   acc2,
		"amount":          100_000,
		"currency":        "UZS",
	}, user1.AccessToken)

	if rec.Code == fiber.StatusCreated {
		t.Error("expected transfer to fail with insufficient funds")
	}

	// Verify balances unchanged
	bal1 := getAccountBalance(t, user1.AccessToken, acc1)
	if bal1 != 10_000 {
		t.Errorf("sender balance = %d, want 10000 (unchanged)", bal1)
	}
	bal2 := getAccountBalance(t, user2.AccessToken, acc2)
	if bal2 != 0 {
		t.Errorf("receiver balance = %d, want 0 (unchanged)", bal2)
	}
}

// TestTransferFlow_TransferToNonExistentAccount tests transfer to a non-existent account.
func TestTransferFlow_TransferToNonExistentAccount(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("xfer-noexist"), "Password123!", "NoExist")
	accID := createAccountWithDeposit(t, user.AccessToken, "UZS", 100_000)

	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": accID,
		"to_account_id":   "00000000-0000-0000-0000-000000000000",
		"amount":          10_000,
		"currency":        "UZS",
	}, user.AccessToken)

	if rec.Code == fiber.StatusCreated {
		t.Error("expected transfer to non-existent account to fail")
	}
}

// TestTransferFlow_UnauthorizedTransfer tests that a user cannot transfer from another user's account.
func TestTransferFlow_UnauthorizedTransfer(t *testing.T) {
	owner := registerAndLogin(t, uniqueEmail("xfer-owner"), "Password123!", "Owner")
	attacker := registerAndLogin(t, uniqueEmail("xfer-attacker"), "Password123!", "Attacker")

	ownerAcc := createAccountWithDeposit(t, owner.AccessToken, "UZS", 500_000)
	attackerAcc := createAccountWithDeposit(t, attacker.AccessToken, "UZS", 0)

	// Attacker tries to transfer from owner's account
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": ownerAcc,
		"to_account_id":   attackerAcc,
		"amount":          100_000,
		"currency":        "UZS",
	}, attacker.AccessToken)

	if rec.Code == fiber.StatusCreated {
		t.Error("expected unauthorized transfer from another user's account to fail")
	}
}

// TestTransferFlow_TransferWithoutAuth tests that transfer requires authentication.
func TestTransferFlow_TransferWithoutAuth(t *testing.T) {
	rec := doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": "some-id",
		"to_account_id":   "other-id",
		"amount":          10_000,
		"currency":        "UZS",
	}, "")

	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

// TestTransferFlow_TransferToClosedAccount tests that transfer to a closed account fails.
func TestTransferFlow_TransferToClosedAccount(t *testing.T) {
	user1 := registerAndLogin(t, uniqueEmail("xfer-closed1"), "Password123!", "Closed1")
	user2 := registerAndLogin(t, uniqueEmail("xfer-closed2"), "Password123!", "Closed2")

	acc1 := createAccountWithDeposit(t, user1.AccessToken, "UZS", 500_000)
	acc2 := createAccountWithDeposit(t, user2.AccessToken, "UZS", 0)

	// Close receiver's account
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/close", map[string]interface{}{
		"account_id": acc2,
	}, user2.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Try to transfer to closed account
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": acc1,
		"to_account_id":   acc2,
		"amount":          50_000,
		"currency":        "UZS",
	}, user1.AccessToken)

	if rec.Code == fiber.StatusCreated {
		t.Error("expected transfer to closed account to fail")
	}
}

// TestTransferFlow_TransferFromClosedAccount tests that transfer from a closed account fails.
func TestTransferFlow_TransferFromClosedAccount(t *testing.T) {
	user1 := registerAndLogin(t, uniqueEmail("xfer-fromc1"), "Password123!", "FromC1")
	user2 := registerAndLogin(t, uniqueEmail("xfer-fromc2"), "Password123!", "FromC2")

	acc1 := createAccountWithDeposit(t, user1.AccessToken, "UZS", 500_000)
	acc2 := createAccountWithDeposit(t, user2.AccessToken, "UZS", 0)

	// Close sender's account
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/close", map[string]interface{}{
		"account_id": acc1,
	}, user1.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Try to transfer from closed account
	rec = doHMACRequest(t, fiber.MethodPost, "/api/v1/transfers/send", map[string]interface{}{
		"from_account_id": acc1,
		"to_account_id":   acc2,
		"amount":          50_000,
		"currency":        "UZS",
	}, user1.AccessToken)

	if rec.Code == fiber.StatusCreated {
		t.Error("expected transfer from closed account to fail")
	}
}
