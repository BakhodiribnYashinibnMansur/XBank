package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX — common interface that both *pgxpool.Pool and pgx.Tx satisfy.
// Repositories use this instead of concrete pool/tx types.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txKey struct{}
type rlsUserKey struct{}

// WithRLSUser injects the user ID into context for Row-Level Security
func WithRLSUser(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, rlsUserKey{}, userID)
}

// injectTx stores the transaction in context.
func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// ExtractDBTX retrieves the transaction from context.
// If no transaction is present, it returns the fallback (typically *pgxpool.Pool).
func ExtractDBTX(ctx context.Context, fallback DBTX) DBTX {
	if tx, ok := ctx.Value(txKey{}).(DBTX); ok {
		return tx
	}
	return fallback
}

// TxManager implements shared.TxManager using pgxpool.Pool.
type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// WithTx executes fn within a database transaction (READ COMMITTED).
func (m *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.withTxOptions(ctx, pgx.TxOptions{}, fn)
}

// WithSerializableTx executes fn within a SERIALIZABLE transaction.
// Prevents phantom reads — required for financial operations (transfers, holds).
// On serialization failure (40001), the caller should retry.
func (m *TxManager) WithSerializableTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.withTxOptions(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	}, fn)
}

func (m *TxManager) withTxOptions(ctx context.Context, opts pgx.TxOptions, fn func(ctx context.Context) error) error {
	tx, err := m.pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set RLS user context if available
	if userID, ok := ctx.Value(rlsUserKey{}).(string); ok && userID != "" {
		if _, err := tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.current_user_id = '%s'", userID)); err != nil {
			return fmt.Errorf("set rls user: %w", err)
		}
	}

	txCtx := injectTx(ctx, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
