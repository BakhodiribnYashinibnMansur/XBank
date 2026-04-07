package domain

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// ScheduledStatus - lifecycle of a scheduled transaction
type ScheduledStatus string

const (
	ScheduledPending   ScheduledStatus = "PENDING"   // waiting for execute_at
	ScheduledExecuted  ScheduledStatus = "EXECUTED"   // successfully processed
	ScheduledFailed    ScheduledStatus = "FAILED"     // execution failed
	ScheduledCancelled ScheduledStatus = "CANCELLED"  // cancelled by user
)

// ScheduledTransfer - a transfer scheduled for future execution.
//
// Flow:
//  1. User creates a scheduled transfer (PENDING)
//  2. pg_cron or background worker picks it up when execute_at <= NOW()
//  3. Transfer is executed via normal transfer flow
//  4. Status updated to EXECUTED or FAILED
type ScheduledTransfer struct {
	ID            string
	UserID        string
	FromAccountID string
	ToAccountID   string
	Amount        domain.Money
	Description   string
	Status        ScheduledStatus
	ExecuteAt     time.Time  // when to execute
	TransferID    string     // ID of the created transfer (after execution)
	FailureReason string
	CreatedAt     time.Time
	ExecutedAt    *time.Time
}

// NewScheduledTransfer - create a new scheduled transfer
func NewScheduledTransfer(userID, fromAccountID, toAccountID string, amount domain.Money, description string, executeAt time.Time) (*ScheduledTransfer, error) {
	if fromAccountID == toAccountID {
		return nil, ErrSameAccount
	}
	if amount.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if executeAt.Before(time.Now()) {
		return nil, fmt.Errorf("execute_at must be in the future")
	}

	id, _ := generateScheduledID()

	return &ScheduledTransfer{
		ID:            id,
		UserID:        userID,
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount,
		Description:   description,
		Status:        ScheduledPending,
		ExecuteAt:     executeAt,
		CreatedAt:     time.Now(),
	}, nil
}

// Cancel - cancel a pending scheduled transfer
func (s *ScheduledTransfer) Cancel() error {
	if s.Status != ScheduledPending {
		return fmt.Errorf("can only cancel pending scheduled transfers")
	}
	s.Status = ScheduledCancelled
	return nil
}

// MarkExecuted - mark as successfully executed
func (s *ScheduledTransfer) MarkExecuted(transferID string) {
	s.Status = ScheduledExecuted
	s.TransferID = transferID
	now := time.Now()
	s.ExecutedAt = &now
}

// MarkFailed - mark as failed
func (s *ScheduledTransfer) MarkFailed(reason string) {
	s.Status = ScheduledFailed
	s.FailureReason = reason
	now := time.Now()
	s.ExecutedAt = &now
}

// ScheduledTransferRepository - persistence for scheduled transfers
type ScheduledTransferRepository interface {
	Create(ctx context.Context, st *ScheduledTransfer) error
	GetByID(ctx context.Context, id string) (*ScheduledTransfer, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*ScheduledTransfer, int64, error)
	Update(ctx context.Context, st *ScheduledTransfer) error
	FetchDue(ctx context.Context, limit int) ([]*ScheduledTransfer, error)
}

func generateScheduledID() (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
