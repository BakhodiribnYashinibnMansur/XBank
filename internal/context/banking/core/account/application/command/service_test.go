package command

import (
	"context"
	"testing"

	account "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/mock"
)

func newTestService() *Service {
	repo := mock.NewAccountRepository()
	eventRepo := mock.NewAccountEventRepository()
	publisher := mock.NewEventPublisher()
	txMgr := mock.NewTxManager()
	auditLog := mock.NewAuditLog()
	return NewService(repo, eventRepo, txMgr, publisher, config.KafkaTopicsConfig{}, auditLog)
}

func TestCreateAccount_Success(t *testing.T) {
	svc := newTestService()

	acc, err := svc.CreateAccount(context.Background(), "user-123", domain.UZS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.ID == "" {
		t.Error("ID should not be empty")
	}
	if acc.Balance.Amount != 0 {
		t.Error("new account balance should be 0")
	}
	if acc.Balance.Currency != domain.UZS {
		t.Errorf("expected currency: UZS, got: %s", acc.Balance.Currency)
	}
}

func TestDeposit_Success(t *testing.T) {
	svc := newTestService()

	acc, _ := svc.CreateAccount(context.Background(), "user-123", domain.UZS)

	updated, err := svc.Deposit(context.Background(), acc.ID, 1000000) // 10000.00 UZS
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Balance.Amount != 1000000 {
		t.Errorf("expected balance: 1000000, got: %d", updated.Balance.Amount)
	}
}

func TestWithdraw_Success(t *testing.T) {
	svc := newTestService()

	acc, _ := svc.CreateAccount(context.Background(), "user-123", domain.UZS)
	svc.Deposit(context.Background(), acc.ID, 1000000)

	updated, err := svc.Withdraw(context.Background(), acc.ID, 300000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Balance.Amount != 700000 {
		t.Errorf("expected balance: 700000, got: %d", updated.Balance.Amount)
	}
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	svc := newTestService()

	acc, _ := svc.CreateAccount(context.Background(), "user-123", domain.UZS)
	svc.Deposit(context.Background(), acc.ID, 100000)

	_, err := svc.Withdraw(context.Background(), acc.ID, 200000)
	if err != domain.ErrInsufficientFunds {
		t.Errorf("expected: %v, got: %v", domain.ErrInsufficientFunds, err)
	}
}

func TestWithdraw_AccountNotFound(t *testing.T) {
	svc := newTestService()

	_, err := svc.Withdraw(context.Background(), "non-existent", 100000)
	if err != account.ErrAccountNotFound {
		t.Errorf("expected: %v, got: %v", account.ErrAccountNotFound, err)
	}
}

func TestCloseAccount_Success(t *testing.T) {
	svc := newTestService()

	acc, _ := svc.CreateAccount(context.Background(), "user-123", domain.UZS)

	err := svc.CloseAccount(context.Background(), acc.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Deposit should fail on closed account
	_, err = svc.Deposit(context.Background(), acc.ID, 100000)
	if err != account.ErrAccountClosed {
		t.Errorf("expected: %v, got: %v", account.ErrAccountClosed, err)
	}
}

func TestCloseAccount_NonZeroBalance(t *testing.T) {
	svc := newTestService()

	acc, _ := svc.CreateAccount(context.Background(), "user-123", domain.UZS)
	svc.Deposit(context.Background(), acc.ID, 100000)

	err := svc.CloseAccount(context.Background(), acc.ID)
	if err != account.ErrBalanceNotZero {
		t.Errorf("expected: %v, got: %v", account.ErrBalanceNotZero, err)
	}
}

func TestListByUserID(t *testing.T) {
	svc := newTestService()

	svc.CreateAccount(context.Background(), "user-123", domain.UZS)
	svc.CreateAccount(context.Background(), "user-123", domain.USD)

	accounts, total, err := svc.ListByUserID(context.Background(), "user-123", 20, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 2 {
		t.Errorf("expected 2 accounts, got: %d", len(accounts))
	}
	if total != 2 {
		t.Errorf("expected total: 2, got: %d", total)
	}
}
