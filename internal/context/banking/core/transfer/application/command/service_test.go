package command

import (
	"context"
	"testing"

	account "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	domainTransfer "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/mock"
)

type testEnv struct {
	svc         *Service
	accountRepo *mock.AccountRepository
}

func setupTransferTest() *testEnv {
	accountRepo := mock.NewAccountRepository()
	transferRepo := mock.NewTransferRepository()
	transferEventRepo := mock.NewTransferEventRepository()
	publisher := mock.NewEventPublisher()
	txMgr := mock.NewTxManager()
	svc := NewService(transferRepo, transferEventRepo, accountRepo, txMgr, publisher, config.KafkaTopicsConfig{})

	// Create two test accounts with balance
	acc1, _ := account.NewAccount("user-1", domain.UZS)
	acc1.Balance = domain.Money{Amount: 1000000, Currency: domain.UZS} // 10000.00 UZS
	accountRepo.Create(context.Background(), acc1)

	acc2, _ := account.NewAccount("user-2", domain.UZS)
	acc2.Balance = domain.Money{Amount: 500000, Currency: domain.UZS} // 5000.00 UZS
	accountRepo.Create(context.Background(), acc2)

	return &testEnv{svc: svc, accountRepo: accountRepo}
}

func getAccountIDs(t *testing.T, repo *mock.AccountRepository) (string, string) {
	t.Helper()
	accounts, _ := repo.ListByUserID(context.Background(), "user-1", 10, 0)
	acc1ID := accounts[0].ID
	accounts, _ = repo.ListByUserID(context.Background(), "user-2", 10, 0)
	acc2ID := accounts[0].ID
	return acc1ID, acc2ID
}

func TestSend_Success(t *testing.T) {
	env := setupTransferTest()
	acc1ID, acc2ID := getAccountIDs(t, env.accountRepo)

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
	from, _ := env.accountRepo.GetByID(context.Background(), acc1ID)
	to, _ := env.accountRepo.GetByID(context.Background(), acc2ID)

	if from.Balance.Amount != 800000 {
		t.Errorf("sender balance expected: 800000, got: %d", from.Balance.Amount)
	}
	if to.Balance.Amount != 700000 {
		t.Errorf("receiver balance expected: 700000, got: %d", to.Balance.Amount)
	}
}

func TestSend_InsufficientFunds(t *testing.T) {
	env := setupTransferTest()
	acc1ID, acc2ID := getAccountIDs(t, env.accountRepo)

	_, err := env.svc.Send(context.Background(), acc1ID, acc2ID, 9999999, domain.UZS, "Too much")
	if err != domain.ErrInsufficientFunds {
		t.Errorf("expected: %v, got: %v", domain.ErrInsufficientFunds, err)
	}
}

func TestSend_SameAccount(t *testing.T) {
	env := setupTransferTest()
	acc1ID, _ := getAccountIDs(t, env.accountRepo)

	_, err := env.svc.Send(context.Background(), acc1ID, acc1ID, 100000, domain.UZS, "Self transfer")
	if err != domainTransfer.ErrSameAccount {
		t.Errorf("expected: %v, got: %v", domainTransfer.ErrSameAccount, err)
	}
}

func TestSend_AccountNotFound(t *testing.T) {
	env := setupTransferTest()
	acc1ID, _ := getAccountIDs(t, env.accountRepo)

	_, err := env.svc.Send(context.Background(), acc1ID, "non-existent", 100000, domain.UZS, "")
	if err != account.ErrAccountNotFound {
		t.Errorf("expected: %v, got: %v", account.ErrAccountNotFound, err)
	}
}

func TestSend_CurrencyMismatch(t *testing.T) {
	env := setupTransferTest()
	acc1ID, acc2ID := getAccountIDs(t, env.accountRepo)

	// Both accounts are UZS, but trying to send USD
	_, err := env.svc.Send(context.Background(), acc1ID, acc2ID, 100000, domain.USD, "Wrong currency")
	if err != domain.ErrCurrencyMismatch {
		t.Errorf("expected: %v, got: %v", domain.ErrCurrencyMismatch, err)
	}
}

func TestSend_ZeroAmount(t *testing.T) {
	env := setupTransferTest()
	acc1ID, acc2ID := getAccountIDs(t, env.accountRepo)

	_, err := env.svc.Send(context.Background(), acc1ID, acc2ID, 0, domain.UZS, "Zero")
	if err != domainTransfer.ErrInvalidAmount {
		t.Errorf("expected: %v, got: %v", domainTransfer.ErrInvalidAmount, err)
	}
}
