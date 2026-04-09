package pgxutil

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// SlowQueryThreshold is the duration after which a query is considered slow.
const SlowQueryThreshold = 100 * time.Millisecond

// DBTX is the common interface that both *pgxpool.Pool and pgx.Tx satisfy.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Exec executes a query with metrics and slow-query tracking.
func Exec(ctx context.Context, db DBTX, operation, sql string, args ...any) (pgconn.CommandTag, error) {
	start := time.Now()
	tag, err := db.Exec(ctx, sql, args...)
	observe(operation, start, err)
	return tag, err
}

// QueryRow executes a query that returns a single row, with metrics.
func QueryRow(ctx context.Context, db DBTX, operation, sql string, args ...any) pgx.Row {
	start := time.Now()
	row := db.QueryRow(ctx, sql, args...)
	observe(operation, start, nil)
	return row
}

// Query executes a query that returns rows, with metrics.
func Query(ctx context.Context, db DBTX, operation, sql string, args ...any) (pgx.Rows, error) {
	start := time.Now()
	rows, err := db.Query(ctx, sql, args...)
	observe(operation, start, err)
	return rows, err
}

// ScanOne scans a single row from the query result into dest using the scanner function.
func ScanOne[T any](ctx context.Context, db DBTX, operation, sql string, scanner func(pgx.Row) (T, error), args ...any) (T, error) {
	start := time.Now()
	row := db.QueryRow(ctx, sql, args...)
	result, err := scanner(row)
	observe(operation, start, err)
	return result, err
}

// ScanMany scans multiple rows from the query result using the scanner function.
func ScanMany[T any](ctx context.Context, db DBTX, operation, sql string, scanner func(pgx.Rows) (T, error), args ...any) ([]T, error) {
	start := time.Now()
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		observe(operation, start, err)
		return nil, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		item, err := scanner(rows)
		if err != nil {
			observe(operation, start, err)
			return nil, err
		}
		results = append(results, item)
	}
	observe(operation, start, rows.Err())
	return results, rows.Err()
}

// IsNotFound checks if the error is pgx.ErrNoRows.
func IsNotFound(err error) bool {
	return err == pgx.ErrNoRows
}

// IsUniqueViolation checks if the error is a PostgreSQL unique constraint violation (23505).
func IsUniqueViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23505"
	}
	return false
}

// WrapNotFound converts pgx.ErrNoRows to a custom error message.
func WrapNotFound(err error, entity string) error {
	if IsNotFound(err) {
		return fmt.Errorf("%s not found", entity)
	}
	return err
}

func observe(operation string, start time.Time, err error) {
	duration := time.Since(start)
	metrics.DBQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
	status := "ok"
	if err != nil {
		status = "error"
	}
	metrics.DBQueryTotal.WithLabelValues(operation, status).Inc()
	if duration > SlowQueryThreshold {
		metrics.DBSlowQueries.WithLabelValues(operation).Inc()
	}
}
