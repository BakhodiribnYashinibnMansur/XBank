package account

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
)

var (
	ErrAccountNotFound = apperror.ErrAccountNotFound
	ErrAccountFrozen   = apperror.ErrAccountFrozen
	ErrAccountClosed   = apperror.ErrAccountClosed
	ErrBalanceNotZero  = apperror.ErrBalanceNotZero
)

// Status - account status
type Status string

const (
	StatusActive Status = "ACTIVE"
	StatusFrozen Status = "FROZEN"
	StatusClosed Status = "CLOSED"
)

// Account - bank account
type Account struct {
	ID            string
	UserID        string
	AccountNumber string        // 16-digit unique number
	Balance       shared.Money  // in tiyin/cent
	Status        Status
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewAccount - create a new account
func NewAccount(userID string, currency shared.Currency) (*Account, error) {
	if userID == "" {
		return nil, apperror.ErrMissingField.WithMessage("user_id cannot be empty")
	}

	accountNumber, err := generateAccountNumber()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Account{
		UserID:        userID,
		AccountNumber: accountNumber,
		Balance:       shared.Money{Amount: 0, Currency: currency},
		Status:        StatusActive,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

// Deposit - add funds to the account
func (a *Account) Deposit(amount shared.Money) error {
	if err := a.checkActive(); err != nil {
		return err
	}

	newBalance, err := a.Balance.Add(amount)
	if err != nil {
		return err
	}

	a.Balance = newBalance
	a.UpdatedAt = time.Now()
	return nil
}

// Withdraw - withdraw funds from the account
func (a *Account) Withdraw(amount shared.Money) error {
	if err := a.checkActive(); err != nil {
		return err
	}

	newBalance, err := a.Balance.Subtract(amount)
	if err != nil {
		return err
	}

	a.Balance = newBalance
	a.UpdatedAt = time.Now()
	return nil
}

// Freeze - freeze the account
func (a *Account) Freeze() error {
	if a.Status == StatusClosed {
		return ErrAccountClosed
	}
	a.Status = StatusFrozen
	a.UpdatedAt = time.Now()
	return nil
}

// Unfreeze - unfreeze the account
func (a *Account) Unfreeze() error {
	if a.Status == StatusClosed {
		return ErrAccountClosed
	}
	a.Status = StatusActive
	a.UpdatedAt = time.Now()
	return nil
}

// Close - close the account (balance must be 0)
func (a *Account) Close() error {
	if !a.Balance.IsZero() {
		return ErrBalanceNotZero
	}
	a.Status = StatusClosed
	a.UpdatedAt = time.Now()
	return nil
}

// checkActive - checks that the account is active
func (a *Account) checkActive() error {
	switch a.Status {
	case StatusFrozen:
		return ErrAccountFrozen
	case StatusClosed:
		return ErrAccountClosed
	}
	return nil
}

// generateAccountNumber - generate a random 16-digit account number
func generateAccountNumber() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%016x", bytes)[:16], nil
}

// Repository - interface for account persistence
type Repository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByIDForUpdate(ctx context.Context, id string) (*Account, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*Account, error)
	CountByUserID(ctx context.Context, userID string) (int64, error)
	Update(ctx context.Context, account *Account) error
}
