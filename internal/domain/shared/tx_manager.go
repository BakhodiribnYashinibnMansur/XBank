package shared

import "context"

// TxManager manages database transactions.
// The application layer uses this to wrap multiple repository calls in a single transaction.
type TxManager interface {
	// WithTx executes fn in a transaction with default isolation (READ COMMITTED).
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error

	// WithSerializableTx executes fn in a SERIALIZABLE transaction.
	// Use for financial operations where phantom reads must be prevented.
	WithSerializableTx(ctx context.Context, fn func(ctx context.Context) error) error
}
