package postgres

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// NewPool - creates a PostgreSQL connection pool
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres pool yaratishda xatolik: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ga ulanib bo'lmadi: %w", err)
	}

	logger.Log.Info("PostgreSQL ga muvaffaqiyatli ulandi",
		zap.String("pool_max_conns", fmt.Sprintf("%d", pool.Config().MaxConns)),
	)
	return pool, nil
}
