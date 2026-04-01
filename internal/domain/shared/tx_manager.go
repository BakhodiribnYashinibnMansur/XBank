package shared

import "context"

// TxManager manages database transactions.
// The application layer uses this to wrap multiple repository calls in a single transaction.
type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
