package ports

import (
	"context"
	"time"
)

// LedgerEntryView is the read-only projection of a ledger entry for cross-BC queries.
type LedgerEntryView struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	TransferID  string    `json:"transfer_id"`
	EntryType   string    `json:"entry_type"` // "DEBIT" or "CREDIT"
	Amount      int64     `json:"amount"`
	Currency    string    `json:"currency"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// LedgerReader provides read-only access to the Ledger BC from other BCs.
type LedgerReader interface {
	ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*LedgerEntryView, error)
	CountByAccountID(ctx context.Context, accountID string) (int64, error)
	BalanceByAccountID(ctx context.Context, accountID string) (int64, error)
}
