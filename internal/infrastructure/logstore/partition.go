// Package logstore provides automatic monthly partitioning for log tables.
// Creates future partitions proactively and drops old ones efficiently.
package logstore

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// PartitionName generates a partition table name for the given month.
// Example: PartitionName("http_request_logs", 2026-04-07) → "http_request_logs_2026_04"
func PartitionName(table string, t time.Time) string {
	return fmt.Sprintf("%s_%d_%02d", table, t.Year(), t.Month())
}

// EnsureFuture creates partitions for the next `ahead` months if they don't exist.
// Safe to call repeatedly — uses IF NOT EXISTS.
func EnsureFuture(ctx context.Context, pool *pgxpool.Pool, table string, ahead int) error {
	now := time.Now()
	for i := 0; i <= ahead; i++ {
		month := now.AddDate(0, i, 0)
		partName := PartitionName(table, month)

		// Partition range: first day of month → first day of next month
		from := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, 0)

		query := fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s')`,
			partName, table,
			from.Format("2006-01-02"),
			to.Format("2006-01-02"),
		)

		if _, err := pool.Exec(ctx, query); err != nil {
			logger.Log.Error("logstore: create partition failed",
				zap.String("partition", partName),
				zap.Error(err),
			)
			return fmt.Errorf("logstore: create partition %s: %w", partName, err)
		}

		logger.Log.Debug("logstore: partition ensured", zap.String("partition", partName))
	}
	return nil
}

// DropOlderThan drops partitions older than the cutoff date.
// Returns the number of partitions dropped. O(1) per partition (DROP TABLE).
func DropOlderThan(ctx context.Context, pool *pgxpool.Pool, table string, cutoff time.Time) (int, error) {
	// List partitions via pg_catalog
	rows, err := pool.Query(ctx,
		`SELECT inhrelid::regclass::text
		 FROM pg_catalog.pg_inherits
		 WHERE inhparent = $1::regclass
		 ORDER BY inhrelid`, table)
	if err != nil {
		return 0, fmt.Errorf("logstore: list partitions: %w", err)
	}
	defer rows.Close()

	dropped := 0
	cutoffName := PartitionName(table, cutoff)

	for rows.Next() {
		var partName string
		if err := rows.Scan(&partName); err != nil {
			return dropped, err
		}

		// Drop only partitions with names lexically before the cutoff
		if partName < cutoffName {
			if _, err := pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", partName)); err != nil {
				logger.Log.Error("logstore: drop partition failed",
					zap.String("partition", partName),
					zap.Error(err),
				)
				continue
			}
			logger.Log.Info("logstore: partition dropped", zap.String("partition", partName))
			dropped++
		}
	}

	return dropped, nil
}
