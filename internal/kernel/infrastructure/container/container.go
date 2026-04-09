package container

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// PostgresContainer provides a PostgreSQL container for integration tests.
type PostgresContainer struct {
	Pool       *pgxpool.Pool
	ConnString string
}

// NewPostgresPool creates a pgxpool.Pool from a DSN (for use in tests with external containers).
func NewPostgresPool(ctx context.Context, dsn string) (*PostgresContainer, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}
	cfg.MaxConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &PostgresContainer{Pool: pool, ConnString: dsn}, nil
}

// Close closes the pool.
func (c *PostgresContainer) Close() {
	if c.Pool != nil {
		c.Pool.Close()
	}
}

// RedisContainer provides a Redis client for integration tests.
type RedisContainer struct {
	Client *redis.Client
}

// NewRedisClient creates a redis.Client from a URL (for use in tests).
func NewRedisClient(ctx context.Context, url string) (*RedisContainer, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisContainer{Client: client}, nil
}

// Close closes the Redis client.
func (c *RedisContainer) Close() {
	if c.Client != nil {
		c.Client.Close()
	}
}
