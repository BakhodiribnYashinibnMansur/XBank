package postgres

import (
	"context"

	account "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// TransferPortAdapter implements ports.AccountTransferPort using the Account domain.
type TransferPortAdapter struct {
	repo account.Repository
}

func NewTransferPortAdapter(repo account.Repository) *TransferPortAdapter {
	return &TransferPortAdapter{repo: repo}
}

func (a *TransferPortAdapter) TransferFunds(ctx context.Context, fromAccountID, toAccountID string, amount int64, currency string) error {
	cur := domain.Currency(currency)
	money, err := domain.NewMoney(amount, cur)
	if err != nil {
		return err
	}

	// Deadlock prevention: lock the smaller ID first
	firstID, secondID := fromAccountID, toAccountID
	if firstID > secondID {
		firstID, secondID = secondID, firstID
	}

	first, err := a.repo.GetByIDForUpdate(ctx, firstID)
	if err != nil {
		return account.ErrAccountNotFound
	}
	second, err := a.repo.GetByIDForUpdate(ctx, secondID)
	if err != nil {
		return account.ErrAccountNotFound
	}

	// Map back to from/to
	var fromAcc, toAcc *account.Account
	if firstID == fromAccountID {
		fromAcc, toAcc = first, second
	} else {
		fromAcc, toAcc = second, first
	}

	// Currency validation
	if fromAcc.Balance.Currency != cur || toAcc.Balance.Currency != cur {
		return domain.ErrCurrencyMismatch
	}

	// Domain operations
	if err := fromAcc.Withdraw(money); err != nil {
		return err
	}
	if err := toAcc.Deposit(money); err != nil {
		return err
	}

	// Persist
	if err := a.repo.Update(ctx, fromAcc); err != nil {
		return err
	}
	return a.repo.Update(ctx, toAcc)
}
