package mock

import (
	"context"

	account "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// MockAccountTransferPort implements ports.AccountTransferPort using mock AccountRepository.
type MockAccountTransferPort struct {
	repo *AccountRepository
}

func NewMockAccountTransferPort(repo *AccountRepository) *MockAccountTransferPort {
	return &MockAccountTransferPort{repo: repo}
}

// SetupTestAccount creates a test account with a given balance (test helper).
func (m *MockAccountTransferPort) SetupTestAccount(userID string, balanceAmount int64, currency domain.Currency) string {
	acc, _ := account.NewAccount(userID, currency)
	acc.Balance = domain.Money{Amount: balanceAmount, Currency: currency}
	m.repo.Create(context.Background(), acc)
	return acc.ID
}

// GetBalance returns account balance for test assertions.
func (m *MockAccountTransferPort) GetBalance(id string) int64 {
	acc, err := m.repo.GetByID(context.Background(), id)
	if err != nil {
		return -1
	}
	return acc.Balance.Amount
}

func (m *MockAccountTransferPort) TransferFunds(ctx context.Context, fromAccountID, toAccountID string, amount int64, currency string) error {
	cur := domain.Currency(currency)
	money, err := domain.NewMoney(amount, cur)
	if err != nil {
		return err
	}

	firstID, secondID := fromAccountID, toAccountID
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}

	first, err := m.repo.GetByIDForUpdate(ctx, firstID)
	if err != nil {
		return account.ErrAccountNotFound
	}
	second, err := m.repo.GetByIDForUpdate(ctx, secondID)
	if err != nil {
		return account.ErrAccountNotFound
	}

	var fromAcc, toAcc *account.Account
	if firstID == fromAccountID {
		fromAcc, toAcc = first, second
	} else {
		fromAcc, toAcc = second, first
	}

	if fromAcc.Balance.Currency != cur || toAcc.Balance.Currency != cur {
		return domain.ErrCurrencyMismatch
	}

	if err := fromAcc.Withdraw(money); err != nil {
		return err
	}
	if err := toAcc.Deposit(money); err != nil {
		return err
	}

	if err := m.repo.Update(ctx, fromAcc); err != nil {
		return err
	}
	return m.repo.Update(ctx, toAcc)
}
