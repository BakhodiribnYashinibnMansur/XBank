package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// RateLimiter - Redis-based sliding window rate limiter
type RateLimiter struct {
	client *goredis.Client
}

func NewRateLimiter(client *goredis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow - check if a request is allowed for the given key
// Returns (allowed, remaining, error)
func (r *RateLimiter) Allow(ctx context.Context, key string, maxRequests int, window time.Duration) (bool, int, error) {
	now := time.Now().UnixMilli()
	windowStart := now - window.Milliseconds()
	redisKey := "ratelimit:" + key

	pipe := r.client.Pipeline()

	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart))

	// Count current entries
	countCmd := pipe.ZCard(ctx, redisKey)

	// Add current request
	pipe.ZAdd(ctx, redisKey, goredis.Z{Score: float64(now), Member: now})

	// Set TTL on the key
	pipe.Expire(ctx, redisKey, window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return true, maxRequests, err // fail open
	}

	count := int(countCmd.Val())
	remaining := maxRequests - count
	if remaining < 0 {
		remaining = 0
	}

	return count < int64ToInt(int64(maxRequests)), remaining, nil
}

func int64ToInt(n int64) int {
	return int(n)
}
