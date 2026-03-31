package transfer

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/account"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/transfer"
)

type Service struct {
	transferRepo transfer.Repository
	accountRepo  account.Repository
	txManager    shared.TxManager
}

func NewService(transferRepo transfer.Repository, accountRepo account.Repository, txManager shared.TxManager) *Service {
	return &Service{
		transferRepo: transferRepo,
		accountRepo:  accountRepo,
		txManager:    txManager,
	}
}

// Send - transfer funds between accounts within a single transaction
func (s *Service) Send(ctx context.Context, fromAccountID, toAccountID string, amount int64, currency shared.Currency, description string) (*transfer.Transfer, error) {
	money, err := shared.NewMoney(amount, currency)
	if err != nil {
		return nil, err
	}

	tr, err := transfer.NewTransfer(fromAccountID, toAccountID, money, description)
	if err != nil {
		return nil, err
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Deadlock prevention: lock the smaller ID first
		firstID, secondID := fromAccountID, toAccountID
		if firstID > secondID {
			firstID, secondID = secondID, firstID
		}

		first, err := s.accountRepo.GetByIDForUpdate(txCtx, firstID)
		if err != nil {
			return account.ErrAccountNotFound
		}

		second, err := s.accountRepo.GetByIDForUpdate(txCtx, secondID)
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
		if fromAcc.Balance.Currency != currency || toAcc.Balance.Currency != currency {
			return shared.ErrCurrencyMismatch
		}

		// Domain operations (checks active status, sufficient funds, etc.)
		if err := fromAcc.Withdraw(money); err != nil {
			return err
		}
		if err := toAcc.Deposit(money); err != nil {
			return err
		}

		// Persist (each repo does a single query)
		if err := s.accountRepo.Update(txCtx, fromAcc); err != nil {
			return err
		}
		if err := s.accountRepo.Update(txCtx, toAcc); err != nil {
			return err
		}

		tr.Complete()
		return s.transferRepo.Create(txCtx, tr)
	})

	if err != nil {
		tr.Fail(err.Error())
		return tr, err
	}

	return tr, nil
}

// GetByID - get a transfer by ID
func (s *Service) GetByID(ctx context.Context, id string) (*transfer.Transfer, error) {
	return s.transferRepo.GetByID(ctx, id)
}

// ListByAccountID - get transfers for an account with pagination
func (s *Service) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*transfer.Transfer, int64, error) {
	transfers, err := s.transferRepo.ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.transferRepo.CountByAccountID(ctx, accountID)
	if err != nil {
		return nil, 0, err
	}
	return transfers, total, nil
}
