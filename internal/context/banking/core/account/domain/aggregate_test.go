package domain

import (
	"testing"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
)

func TestNewAccount(t *testing.T) {
	acc, err := NewAccount("user-123", shared.UZS)
	if err != nil {
		t.Fatalf("Kutilmagan xatolik: %v", err)
	}
	if acc.UserID != "user-123" {
		t.Errorf("UserID kutilgan: user-123, kelgan: %s", acc.UserID)
	}
	if acc.Status != StatusActive {
		t.Errorf("Status ACTIVE bo'lishi kerak, kelgan: %s", acc.Status)
	}
	if !acc.Balance.IsZero() {
		t.Error("Yangi hisob balans 0 bo'lishi kerak")
	}
	if len(acc.AccountNumber) != 16 {
		t.Errorf("Account number 16 ta belgi bo'lishi kerak, kelgan: %d", len(acc.AccountNumber))
	}
}

func TestAccount_Deposit(t *testing.T) {
	acc, _ := NewAccount("user-123", shared.UZS)
	amount, _ := shared.NewMoney(1500050, shared.UZS) // 15000.50 UZS

	if err := acc.Deposit(amount); err != nil {
		t.Fatalf("Deposit xatolik: %v", err)
	}
	if acc.Balance.Amount != 1500050 {
		t.Errorf("Balance kutilgan: 1500050, kelgan: %d", acc.Balance.Amount)
	}
}

func TestAccount_Withdraw(t *testing.T) {
	acc, _ := NewAccount("user-123", shared.UZS)
	deposit, _ := shared.NewMoney(1000000, shared.UZS)
	acc.Deposit(deposit)

	withdraw, _ := shared.NewMoney(300000, shared.UZS)
	if err := acc.Withdraw(withdraw); err != nil {
		t.Fatalf("Withdraw xatolik: %v", err)
	}
	if acc.Balance.Amount != 700000 {
		t.Errorf("Balance kutilgan: 700000, kelgan: %d", acc.Balance.Amount)
	}
}

func TestAccount_Withdraw_InsufficientFunds(t *testing.T) {
	acc, _ := NewAccount("user-123", shared.UZS)
	deposit, _ := shared.NewMoney(100000, shared.UZS)
	acc.Deposit(deposit)

	withdraw, _ := shared.NewMoney(200000, shared.UZS)
	err := acc.Withdraw(withdraw)
	if err != shared.ErrInsufficientFunds {
		t.Errorf("Kutilgan: %v, kelgan: %v", shared.ErrInsufficientFunds, err)
	}
}

func TestAccount_Freeze_Unfreeze(t *testing.T) {
	acc, _ := NewAccount("user-123", shared.UZS)

	acc.Freeze()
	if acc.Status != StatusFrozen {
		t.Error("Status FROZEN bo'lishi kerak")
	}

	// Muzlatilgan hisobga deposit bo'lmaydi
	amount, _ := shared.NewMoney(100000, shared.UZS)
	if err := acc.Deposit(amount); err != ErrAccountFrozen {
		t.Errorf("Frozen hisobda ErrAccountFrozen kutilgan, kelgan: %v", err)
	}

	acc.Unfreeze()
	if acc.Status != StatusActive {
		t.Error("Status ACTIVE bo'lishi kerak")
	}
}

func TestAccount_Close(t *testing.T) {
	acc, _ := NewAccount("user-123", shared.UZS)

	// Balans 0 bo'lsa yopish mumkin
	if err := acc.Close(); err != nil {
		t.Fatalf("Close xatolik: %v", err)
	}
	if acc.Status != StatusClosed {
		t.Error("Status CLOSED bo'lishi kerak")
	}
}

func TestAccount_Close_NonZeroBalance(t *testing.T) {
	acc, _ := NewAccount("user-123", shared.UZS)
	deposit, _ := shared.NewMoney(100000, shared.UZS)
	acc.Deposit(deposit)

	err := acc.Close()
	if err != ErrBalanceNotZero {
		t.Errorf("Kutilgan: %v, kelgan: %v", ErrBalanceNotZero, err)
	}
}
