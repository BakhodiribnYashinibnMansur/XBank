package redis

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// NewClient - create and connect a Redis client
func NewClient(ctx context.Context, url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("redis URL parse error: %w", err)
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping error: %w", err)
	}

	logger.Log.Info("Redis connected", zap.String("addr", opts.Addr))
	return client, nil
}
