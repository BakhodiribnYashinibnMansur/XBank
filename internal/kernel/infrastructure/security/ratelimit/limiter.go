package ratelimit

import (
	"context"
	"fmt"
	"math"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Operation identifies an auth-related operation for rate limiting.
type Operation string

const (
	OpLogin         Operation = "login"
	OpPasswordReset Operation = "password_reset"
	OpOTPVerify     Operation = "otp_verify"
	OpRegister      Operation = "register"
)

// Config defines rate limiting parameters for a single operation.
type Config struct {
	MaxAttempts  int
	Window       time.Duration
	LockDuration time.Duration
	Progressive  bool // enable exponential backoff on lock duration
}

// Limiter provides auth-specific rate limiting with per-operation configs.
type Limiter struct {
	client  *goredis.Client
	configs map[Operation]Config
}

// NewLimiter creates a Limiter backed by Redis.
func NewLimiter(client *goredis.Client) *Limiter {
	return &Limiter{
		client:  client,
		configs: make(map[Operation]Config),
	}
}

// Configure sets the rate limit config for a specific operation.
func (l *Limiter) Configure(op Operation, cfg Config) {
	l.configs[op] = cfg
}

// Allow checks whether the identifier is allowed to attempt the given operation.
// Returns (allowed, retryAfter).
func (l *Limiter) Allow(ctx context.Context, op Operation, identifier string) (bool, time.Duration, error) {
	cfg, ok := l.configs[op]
	if !ok {
		return true, 0, nil // unconfigured operations are allowed
	}

	lockKey := lockKey(op, identifier)

	// Check if locked
	ttl, err := l.client.TTL(ctx, lockKey).Result()
	if err != nil && err != goredis.Nil {
		return true, 0, nil // fail open
	}
	if ttl > 0 {
		return false, ttl, nil
	}

	// Check attempt count
	attemptsKey := attemptsKey(op, identifier)
	count, err := l.client.Get(ctx, attemptsKey).Int()
	if err != nil && err != goredis.Nil {
		return true, 0, nil // fail open
	}

	return count < cfg.MaxAttempts, 0, nil
}

// RecordFailure increments the failure counter. If the max is reached, locks the identifier.
// Returns (locked, error).
func (l *Limiter) RecordFailure(ctx context.Context, op Operation, identifier string) (bool, error) {
	cfg, ok := l.configs[op]
	if !ok {
		return false, nil
	}

	attemptsKey := attemptsKey(op, identifier)

	pipe := l.client.Pipeline()
	incrCmd := pipe.Incr(ctx, attemptsKey)
	pipe.Expire(ctx, attemptsKey, cfg.Window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := int(incrCmd.Val())
	if count >= cfg.MaxAttempts {
		lockDuration := cfg.LockDuration
		if cfg.Progressive {
			// Exponential backoff: lockDuration * 2^(violations-1)
			violations := (count - cfg.MaxAttempts) / cfg.MaxAttempts
			multiplier := math.Pow(2, float64(violations))
			lockDuration = time.Duration(float64(lockDuration) * multiplier)
			if lockDuration > 24*time.Hour {
				lockDuration = 24 * time.Hour
			}
		}
		lockKey := lockKey(op, identifier)
		l.client.Set(ctx, lockKey, "1", lockDuration)
		return true, nil
	}

	return false, nil
}

// Reset clears the failure counter and lock for an identifier.
func (l *Limiter) Reset(ctx context.Context, op Operation, identifier string) {
	l.client.Del(ctx, attemptsKey(op, identifier), lockKey(op, identifier))
}

func attemptsKey(op Operation, id string) string {
	return fmt.Sprintf("auth_rl:%s:%s", op, id)
}

func lockKey(op Operation, id string) string {
	return fmt.Sprintf("auth_lock:%s:%s", op, id)
}
