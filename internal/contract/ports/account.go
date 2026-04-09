package ports

import "context"

// AccountView is the read-only projection of an account for cross-BC queries.
type AccountView struct {
	ID            string `json:"id"`
	UserID        string `json:"user_id"`
	AccountNumber string `json:"account_number"`
	Currency      string `json:"currency"`
	Balance       int64  `json:"balance"`
	Status        string `json:"status"`
}

// AccountReader provides read-only access to the Account BC from other BCs.
// This is the Anti-Corruption Layer (ACL) — consuming BCs depend on this interface,
// not on the Account domain directly.
type AccountReader interface {
	GetByID(ctx context.Context, id string) (*AccountView, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*AccountView, error)
	CountByUserID(ctx context.Context, userID string) (int64, error)
}

// AccountWriter provides write access to the Account BC for saga orchestration.
type AccountWriter interface {
	GetByIDForUpdate(ctx context.Context, id string) (*AccountView, error)
	Credit(ctx context.Context, accountID string, amount int64) error
	Debit(ctx context.Context, accountID string, amount int64) error
}

// AccountTransferPort provides the transactional transfer operation.
// The Account BC implements this so that the Transfer BC doesn't need
// direct access to Account domain entities.
type AccountTransferPort interface {
	// TransferFunds atomically debits fromAccount and credits toAccount.
	// It validates currency match, sufficient funds, and active status.
	TransferFunds(ctx context.Context, fromAccountID, toAccountID string, amount int64, currency string) error
}
