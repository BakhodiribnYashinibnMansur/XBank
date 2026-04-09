package apikeythrottle

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Tier defines rate limit parameters for an API key tier.
type Tier struct {
	MaxRequests int
	Window      time.Duration
}

// Throttle provides per-API-key rate limiting using Redis sliding windows.
type Throttle struct {
	client      *goredis.Client
	tiers       map[string]Tier
	defaultTier Tier
}

// NewThrottle creates a throttle with a default tier.
func NewThrottle(client *goredis.Client, defaultTier Tier) *Throttle {
	return &Throttle{
		client:      client,
		tiers:       make(map[string]Tier),
		defaultTier: defaultTier,
	}
}

// SetTier registers a named tier with specific limits.
func (t *Throttle) SetTier(name string, tier Tier) {
	t.tiers[name] = tier
}

// Allow checks whether a request from the given API key is allowed.
// Returns (allowed, remaining requests, error).
func (t *Throttle) Allow(ctx context.Context, keyID string, tierName string) (bool, int, error) {
	tier := t.defaultTier
	if named, ok := t.tiers[tierName]; ok {
		tier = named
	}

	now := time.Now().UnixMilli()
	windowStart := now - tier.Window.Milliseconds()
	redisKey := fmt.Sprintf("apikey_rl:%s", keyID)

	pipe := t.client.Pipeline()
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart))
	countCmd := pipe.ZCard(ctx, redisKey)
	pipe.ZAdd(ctx, redisKey, goredis.Z{Score: float64(now), Member: now})
	pipe.Expire(ctx, redisKey, tier.Window)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return true, tier.MaxRequests, nil // fail open
	}

	count := int(countCmd.Val())
	remaining := tier.MaxRequests - count
	if remaining < 0 {
		remaining = 0
	}

	return count < tier.MaxRequests, remaining, nil
}
