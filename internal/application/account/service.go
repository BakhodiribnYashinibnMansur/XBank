package account

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/account"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
)

type Service struct {
	repo      account.Repository
	txManager shared.TxManager
}

func NewService(repo account.Repository, txManager shared.TxManager) *Service {
	return &Service{repo: repo, txManager: txManager}
}

// CreateAccount - open a new account
func (s *Service) CreateAccount(ctx context.Context, userID string, currency shared.Currency) (*account.Account, error) {
	acc, err := account.NewAccount(userID, currency)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, acc); err != nil {
		return nil, err
	}

	return acc, nil
}

// GetByID - get an account by ID
func (s *Service) GetByID(ctx context.Context, id string) (*account.Account, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByUserID - get accounts for a user with pagination
func (s *Service) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*account.Account, int64, error) {
	accounts, err := s.repo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return accounts, total, nil
}

// Deposit - add funds to an account (transactional, race-condition safe)
func (s *Service) Deposit(ctx context.Context, accountID string, amount int64) (*account.Account, error) {
	var result *account.Account

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		acc, err := s.repo.GetByIDForUpdate(txCtx, accountID)
		if err != nil {
			return err
		}

		money, err := shared.NewMoney(amount, acc.Balance.Currency)
		if err != nil {
			return err
		}

		if err := acc.Deposit(money); err != nil {
			return err
		}

		if err := s.repo.Update(txCtx, acc); err != nil {
			return err
		}

		result = acc
		return nil
	})

	return result, err
}

// Withdraw - withdraw funds from an account (transactional, race-condition safe)
func (s *Service) Withdraw(ctx context.Context, accountID string, amount int64) (*account.Account, error) {
	var result *account.Account

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		acc, err := s.repo.GetByIDForUpdate(txCtx, accountID)
		if err != nil {
			return err
		}

		money, err := shared.NewMoney(amount, acc.Balance.Currency)
		if err != nil {
			return err
		}

		if err := acc.Withdraw(money); err != nil {
			return err
		}

		if err := s.repo.Update(txCtx, acc); err != nil {
			return err
		}

		result = acc
		return nil
	})

	return result, err
}

// CloseAccount - close an account (transactional)
func (s *Service) CloseAccount(ctx context.Context, accountID string) error {
	return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		acc, err := s.repo.GetByIDForUpdate(txCtx, accountID)
		if err != nil {
			return err
		}

		if err := acc.Close(); err != nil {
			return err
		}

		return s.repo.Update(txCtx, acc)
	})
}
