package metrics

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const slowQueryThreshold = 100 * time.Millisecond

// StartDBPoolCollector periodically collects pgxpool.Stat() and updates Prometheus gauges.
// The goroutine stops when ctx is cancelled.
func StartDBPoolCollector(ctx context.Context, pool *pgxpool.Pool, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stat := pool.Stat()
				DBPoolSize.WithLabelValues("total").Set(float64(stat.TotalConns()))
				DBPoolSize.WithLabelValues("idle").Set(float64(stat.IdleConns()))
				DBPoolSize.WithLabelValues("acquired").Set(float64(stat.AcquiredConns()))
			}
		}
	}()
}

// ObserveQuery records query duration, increments counters, and logs slow queries.
// Usage:
//
//	defer metrics.ObserveQuery("AccountRepo.GetByID", time.Now(), err)
func ObserveQuery(operation string, start time.Time, err error) {
	duration := time.Since(start)
	seconds := duration.Seconds()

	DBQueryDuration.WithLabelValues(operation).Observe(seconds)

	status := "ok"
	if err != nil {
		status = "error"
	}
	DBQueryTotal.WithLabelValues(operation, status).Inc()

	if duration >= slowQueryThreshold {
		DBSlowQueries.WithLabelValues(operation).Inc()
		logger.Log.Warn("slow query detected",
			zap.String("operation", operation),
			zap.Duration("duration", duration),
			zap.Bool("error", err != nil),
		)
	}
}
