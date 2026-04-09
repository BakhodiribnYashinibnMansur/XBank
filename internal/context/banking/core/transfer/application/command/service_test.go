package command

import (
	"context"
	"testing"

	domainTransfer "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/mock"
)

type testEnv struct {
	svc         *Service
	accountPort *mock.MockAccountTransferPort
}

func setupTransferTest() *testEnv {
	accountRepo := mock.NewAccountRepository()
	accountPort := mock.NewMockAccountTransferPort(accountRepo)
	transferRepo := mock.NewTransferRepository()
	transferEventRepo := mock.NewTransferEventRepository()
	publisher := mock.NewEventPublisher()
	txMgr := mock.NewTxManager()
	svc := NewService(transferRepo, transferEventRepo, accountPort, txMgr, publisher, config.KafkaTopicsConfig{})

	return &testEnv{svc: svc, accountPort: accountPort}
}

func TestSend_Success(t *testing.T) {
	env := setupTransferTest()
	acc1ID := env.accountPort.SetupTestAccount("user-1", 1000000, domain.UZS)
	acc2ID := env.accountPort.SetupTestAccount("user-2", 500000, domain.UZS)

	tr, err := env.svc.Send(context.Background(), acc1ID, acc2ID, 200000, domain.UZS, "Test transfer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr.Status != domainTransfer.StatusCompleted {
		t.Errorf("expected status: COMPLETED, got: %s", tr.Status)
	}
	if tr.ID == "" {
		t.Error("transfer ID should not be empty")
	}

	// Verify balances
	if bal := env.accountPort.GetBalance(acc1ID); bal != 800000 {
		t.Errorf("sender balance expected: 800000, got: %d", bal)
	}
	if bal := env.accountPort.GetBalance(acc2ID); bal != 700000 {
		t.Errorf("receiver balance expected: 700000, got: %d", bal)
	}
}

func TestSend_InsufficientFunds(t *testing.T) {
	env := setupTransferTest()
	acc1ID := env.accountPort.SetupTestAccount("user-1", 1000000, domain.UZS)
	acc2ID := env.accountPort.SetupTestAccount("user-2", 500000, domain.UZS)

	_, err := env.svc.Send(context.Background(), acc1ID, acc2ID, 9999999, domain.UZS, "Too much")
	if err != domain.ErrInsufficientFunds {
		t.Errorf("expected: %v, got: %v", domain.ErrInsufficientFunds, err)
	}
}

func TestSend_SameAccount(t *testing.T) {
	env := setupTransferTest()
	acc1ID := env.accountPort.SetupTestAccount("user-1", 1000000, domain.UZS)

	_, err := env.svc.Send(context.Background(), acc1ID, acc1ID, 100000, domain.UZS, "Self transfer")
	if err != domainTransfer.ErrSameAccount {
		t.Errorf("expected: %v, got: %v", domainTransfer.ErrSameAccount, err)
	}
}

func TestSend_AccountNotFound(t *testing.T) {
	env := setupTransferTest()
	acc1ID := env.accountPort.SetupTestAccount("user-1", 1000000, domain.UZS)

	_, err := env.svc.Send(context.Background(), acc1ID, "non-existent", 100000, domain.UZS, "")
	if err == nil {
		t.Error("expected error for non-existent account")
	}
}

func TestSend_CurrencyMismatch(t *testing.T) {
	env := setupTransferTest()
	acc1ID := env.accountPort.SetupTestAccount("user-1", 1000000, domain.UZS)
	acc2ID := env.accountPort.SetupTestAccount("user-2", 500000, domain.UZS)

	// Both accounts are UZS, but trying to send USD
	_, err := env.svc.Send(context.Background(), acc1ID, acc2ID, 100000, domain.USD, "Wrong currency")
	if err != domain.ErrCurrencyMismatch {
		t.Errorf("expected: %v, got: %v", domain.ErrCurrencyMismatch, err)
	}
}

func TestSend_ZeroAmount(t *testing.T) {
	env := setupTransferTest()
	acc1ID := env.accountPort.SetupTestAccount("user-1", 1000000, domain.UZS)
	acc2ID := env.accountPort.SetupTestAccount("user-2", 500000, domain.UZS)

	_, err := env.svc.Send(context.Background(), acc1ID, acc2ID, 0, domain.UZS, "Zero")
	if err != domainTransfer.ErrInvalidAmount {
		t.Errorf("expected: %v, got: %v", domainTransfer.ErrInvalidAmount, err)
	}
}
