package e2e_test

import (
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestAccountFlow_CreateGetListBalance tests the full account lifecycle:
// create -> get -> list -> deposit -> withdraw -> check balance -> history
func TestAccountFlow_CreateGetListBalance(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("acc-flow"), "Password123!", "AccFlow")

	// Step 1: Create a UZS account
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/create", map[string]string{
		"currency": "UZS",
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusCreated)

	var account struct {
		ID            string `json:"id"`
		AccountNumber string `json:"account_number"`
		Balance       int64  `json:"balance"`
		Currency      string `json:"currency"`
		Status        string `json:"status"`
	}
	parseResponse(t, rec, &account)

	if account.ID == "" {
		t.Fatal("expected account ID")
	}
	if account.Balance != 0 {
		t.Errorf("initial balance = %d, want 0", account.Balance)
	}
	if account.Currency != "UZS" {
		t.Errorf("currency = %q, want UZS", account.Currency)
	}
	if account.Status != "ACTIVE" {
		t.Errorf("status = %q, want ACTIVE", account.Status)
	}
	if account.AccountNumber == "" {
		t.Error("expected account number to be set")
	}

	// Step 2: Get account by ID
	rec = doRequest(t, fiber.MethodGet, "/api/v1/accounts/get?id="+account.ID, nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	var fetched struct {
		ID       string `json:"id"`
		Balance  int64  `json:"balance"`
		Currency string `json:"currency"`
	}
	parseResponse(t, rec, &fetched)
	if fetched.ID != account.ID {
		t.Errorf("fetched ID = %q, want %q", fetched.ID, account.ID)
	}

	// Step 3: List user's accounts
	rec = doRequest(t, fiber.MethodGet, "/api/v1/accounts/list", nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Step 4: Deposit money
	rec = doRequest(t, fiber.MethodPost, "/api/v1/accounts/deposit", map[string]interface{}{
		"account_id": account.ID,
		"amount":     1_500_000,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	var afterDeposit struct {
		Balance int64 `json:"balance"`
	}
	parseResponse(t, rec, &afterDeposit)
	if afterDeposit.Balance != 1_500_000 {
		t.Errorf("balance after deposit = %d, want 1500000", afterDeposit.Balance)
	}

	// Step 5: Withdraw money
	rec = doRequest(t, fiber.MethodPost, "/api/v1/accounts/withdraw", map[string]interface{}{
		"account_id": account.ID,
		"amount":     500_000,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	var afterWithdraw struct {
		Balance int64 `json:"balance"`
	}
	parseResponse(t, rec, &afterWithdraw)
	if afterWithdraw.Balance != 1_000_000 {
		t.Errorf("balance after withdraw = %d, want 1000000", afterWithdraw.Balance)
	}

	// Step 6: Account history
	rec = doRequest(t, fiber.MethodGet, "/api/v1/accounts/history?id="+account.ID, nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestAccountFlow_MultiCurrency tests creating accounts in different currencies.
func TestAccountFlow_MultiCurrency(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("multi-curr"), "Password123!", "MultiCur")

	currencies := []string{"UZS", "USD", "EUR"}

	for _, curr := range currencies {
		t.Run(curr, func(t *testing.T) {
			rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/create", map[string]string{
				"currency": curr,
			}, user.AccessToken)
			expectStatus(t, rec, fiber.StatusCreated)

			var acc struct {
				ID       string `json:"id"`
				Currency string `json:"currency"`
				Status   string `json:"status"`
			}
			parseResponse(t, rec, &acc)

			if acc.Currency != curr {
				t.Errorf("currency = %q, want %q", acc.Currency, curr)
			}
			if acc.Status != "ACTIVE" {
				t.Errorf("status = %q, want ACTIVE", acc.Status)
			}
		})
	}

	// List should show all accounts
	rec := doRequest(t, fiber.MethodGet, "/api/v1/accounts/list", nil, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)
}

// TestAccountFlow_CloseAccount tests account closing.
func TestAccountFlow_CloseAccount(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("close-acc"), "Password123!", "Close")

	accID := createAccountWithDeposit(t, user.AccessToken, "UZS", 0)

	// Close account
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/close", map[string]interface{}{
		"account_id": accID,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Get account -- should show CLOSED status
	rec = doRequest(t, fiber.MethodGet, "/api/v1/accounts/get?id="+accID, nil, user.AccessToken)
	if rec.Code == fiber.StatusOK {
		var closed struct {
			Status string `json:"status"`
		}
		parseResponse(t, rec, &closed)
		if closed.Status != "CLOSED" {
			t.Errorf("closed account status = %q, want CLOSED", closed.Status)
		}
	}
}

// TestAccountFlow_WithdrawExceedingBalance tests that overdraft is rejected.
func TestAccountFlow_WithdrawExceedingBalance(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("overdraft"), "Password123!", "Over")

	accID := createAccountWithDeposit(t, user.AccessToken, "UZS", 10_000)

	// Try to withdraw more than balance
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/withdraw", map[string]interface{}{
		"account_id": accID,
		"amount":     50_000,
	}, user.AccessToken)

	if rec.Code == fiber.StatusOK {
		t.Error("expected overdraft withdrawal to fail")
	}

	// Verify balance unchanged
	balance := getAccountBalance(t, user.AccessToken, accID)
	if balance != 10_000 {
		t.Errorf("balance after failed withdrawal = %d, want 10000", balance)
	}
}

// TestAccountFlow_DepositToClosedAccount tests that deposits to a closed account fail.
func TestAccountFlow_DepositToClosedAccount(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("dep-closed"), "Password123!", "DepClosed")

	accID := createAccountWithDeposit(t, user.AccessToken, "UZS", 0)

	// Close the account
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/close", map[string]interface{}{
		"account_id": accID,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Try to deposit into closed account
	rec = doRequest(t, fiber.MethodPost, "/api/v1/accounts/deposit", map[string]interface{}{
		"account_id": accID,
		"amount":     100_000,
	}, user.AccessToken)

	if rec.Code == fiber.StatusOK {
		t.Error("expected deposit to closed account to fail")
	}
}

// TestAccountFlow_WithdrawFromClosedAccount tests that withdrawal from a closed account fails.
func TestAccountFlow_WithdrawFromClosedAccount(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("wd-closed"), "Password123!", "WdClosed")

	accID := createAccountWithDeposit(t, user.AccessToken, "UZS", 100_000)

	// Close the account
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/close", map[string]interface{}{
		"account_id": accID,
	}, user.AccessToken)
	expectStatus(t, rec, fiber.StatusOK)

	// Try to withdraw from closed account
	rec = doRequest(t, fiber.MethodPost, "/api/v1/accounts/withdraw", map[string]interface{}{
		"account_id": accID,
		"amount":     50_000,
	}, user.AccessToken)

	if rec.Code == fiber.StatusOK {
		t.Error("expected withdrawal from closed account to fail")
	}
}

// TestAccountFlow_UnauthorizedAccess tests that a user cannot access another user's account.
func TestAccountFlow_UnauthorizedAccess(t *testing.T) {
	user1 := registerAndLogin(t, uniqueEmail("acc-owner"), "Password123!", "Owner")
	user2 := registerAndLogin(t, uniqueEmail("acc-intruder"), "Password123!", "Intruder")

	// User1 creates an account
	accID := createAccountWithDeposit(t, user1.AccessToken, "UZS", 500_000)

	// User2 tries to withdraw from user1's account
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/withdraw", map[string]interface{}{
		"account_id": accID,
		"amount":     100_000,
	}, user2.AccessToken)

	if rec.Code == fiber.StatusOK {
		t.Error("expected unauthorized withdrawal to fail")
	}

	// User2 tries to deposit to user1's account
	rec = doRequest(t, fiber.MethodPost, "/api/v1/accounts/deposit", map[string]interface{}{
		"account_id": accID,
		"amount":     100_000,
	}, user2.AccessToken)

	if rec.Code == fiber.StatusOK {
		t.Error("expected unauthorized deposit to fail")
	}

	// User2 tries to close user1's account
	rec = doRequest(t, fiber.MethodPost, "/api/v1/accounts/close", map[string]interface{}{
		"account_id": accID,
	}, user2.AccessToken)

	if rec.Code == fiber.StatusOK {
		t.Error("expected unauthorized close to fail")
	}
}

// TestAccountFlow_DepositNegativeAmount tests that negative deposit is rejected.
func TestAccountFlow_DepositNegativeAmount(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("neg-dep"), "Password123!", "NegDep")
	accID := createAccountWithDeposit(t, user.AccessToken, "UZS", 0)

	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/deposit", map[string]interface{}{
		"account_id": accID,
		"amount":     -100_000,
	}, user.AccessToken)

	if rec.Code == fiber.StatusOK {
		t.Error("expected negative deposit to fail")
	}
}

// TestAccountFlow_DepositZeroAmount tests that zero deposit is rejected.
func TestAccountFlow_DepositZeroAmount(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("zero-dep"), "Password123!", "ZeroDep")
	accID := createAccountWithDeposit(t, user.AccessToken, "UZS", 0)

	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/deposit", map[string]interface{}{
		"account_id": accID,
		"amount":     0,
	}, user.AccessToken)

	if rec.Code == fiber.StatusOK {
		t.Error("expected zero deposit to fail")
	}
}

// TestAccountFlow_MultipleDeposits tests sequential deposits add up correctly.
func TestAccountFlow_MultipleDeposits(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("multi-dep"), "Password123!", "MultiDep")
	accID := createAccountWithDeposit(t, user.AccessToken, "UZS", 0)

	deposits := []int64{100_000, 250_000, 50_000, 300_000}
	var expectedTotal int64

	for _, amount := range deposits {
		expectedTotal += amount
		rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/deposit", map[string]interface{}{
			"account_id": accID,
			"amount":     amount,
		}, user.AccessToken)
		expectStatus(t, rec, fiber.StatusOK)
	}

	balance := getAccountBalance(t, user.AccessToken, accID)
	if balance != expectedTotal {
		t.Errorf("balance = %d, want %d", balance, expectedTotal)
	}
}

// TestAccountFlow_GetNonExistentAccount tests getting an account that does not exist.
func TestAccountFlow_GetNonExistentAccount(t *testing.T) {
	user := registerAndLogin(t, uniqueEmail("no-acc"), "Password123!", "NoAcc")

	rec := doRequest(t, fiber.MethodGet, "/api/v1/accounts/get?id=00000000-0000-0000-0000-000000000000", nil, user.AccessToken)
	if rec.Code == fiber.StatusOK {
		t.Error("expected non-existent account to return error")
	}
}

// TestAccountFlow_CreateWithoutAuth tests that account creation requires auth.
func TestAccountFlow_CreateWithoutAuth(t *testing.T) {
	rec := doRequest(t, fiber.MethodPost, "/api/v1/accounts/create", map[string]string{
		"currency": "UZS",
	}, "")
	if rec.Code != fiber.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
