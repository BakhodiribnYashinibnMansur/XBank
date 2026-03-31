package transfer

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
)

var (
	ErrTransferNotFound = apperror.ErrTransferNotFound
	ErrSameAccount      = apperror.ErrSameAccount
	ErrInvalidAmount    = apperror.ErrInvalidAmount
	ErrTransferFailed   = apperror.ErrTransferFailed
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
)

// Transfer - money transfer between accounts
type Transfer struct {
	ID              string
	FromAccountID   string
	ToAccountID     string
	Amount          shared.Money
	Status          Status
	Description     string
	FailureReason   string
	CreatedAt       time.Time
}

// NewTransfer - create a new transfer (with business validation)
func NewTransfer(fromAccountID, toAccountID string, amount shared.Money, description string) (*Transfer, error) {
	if fromAccountID == toAccountID {
		return nil, ErrSameAccount
	}
	if amount.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	return &Transfer{
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
		Status:        StatusPending,
		Description:   description,
		CreatedAt:     time.Now(),
	}, nil
}

// Complete - mark the transfer as successfully completed
func (t *Transfer) Complete() {
	t.Status = StatusCompleted
}

// Fail - mark the transfer as failed
func (t *Transfer) Fail(reason string) {
	t.Status = StatusFailed
	t.FailureReason = reason
}

// Repository - interface for transfer persistence
type Repository interface {
	Create(ctx context.Context, transfer *Transfer) error
	GetByID(ctx context.Context, id string) (*Transfer, error)
	ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*Transfer, error)
	CountByAccountID(ctx context.Context, accountID string) (int64, error)
}
