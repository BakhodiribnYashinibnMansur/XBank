package ledger

import (
	"context"
	"time"
)

type EntryType string

const (
	Debit  EntryType = "DEBIT"
	Credit EntryType = "CREDIT"
)

// Entry - a single ledger entry (immutable, append-only)
type Entry struct {
	ID         string
	AccountID  string
	TransferID string
	EntryType  EntryType
	Amount     int64  // always positive, type determines direction
	Currency   string
	CreatedAt  time.Time
}

// CreatePair - create debit + credit entries for a transfer
func CreatePair(transferID, fromAccountID, toAccountID string, amount int64, currency string) (debit, credit *Entry) {
	now := time.Now()
	debit = &Entry{
		AccountID:  fromAccountID,
		TransferID: transferID,
		EntryType:  Debit,
		Amount:     amount,
		Currency:   currency,
		CreatedAt:  now,
	}
	credit = &Entry{
		AccountID:  toAccountID,
		TransferID: transferID,
		EntryType:  Credit,
		Amount:     amount,
		Currency:   currency,
		CreatedAt:  now,
	}
	return debit, credit
}

type Repository interface {
	CreatePair(ctx context.Context, debit, credit *Entry) error
	ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*Entry, error)
	CountByAccountID(ctx context.Context, accountID string) (int64, error)
	// SumByAccountID returns (total_credits - total_debits) for verification
	BalanceByAccountID(ctx context.Context, accountID string) (int64, error)
}
