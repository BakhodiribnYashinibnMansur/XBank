package mock

import "context"

// TxManager - no-op transaction manager for testing.
// Simply executes the function directly without wrapping in a real DB transaction.
type TxManager struct{}

func NewTxManager() *TxManager {
	return &TxManager{}
}

func (m *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
