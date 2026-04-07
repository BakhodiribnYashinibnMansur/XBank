package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// ChallengeCache - Redis-backed fast token lookup for step-up auth.
// After a challenge is verified, the token is cached here for quick
// validation by the RequireChallenge middleware.
//
// Key pattern: challenge_token:{token} → user_id
type ChallengeCache struct {
	client *redis.Client
}

func NewChallengeCache(client *redis.Client) *ChallengeCache {
	return &ChallengeCache{client: client}
}

// SetToken - cache a verified challenge token
func (c *ChallengeCache) SetToken(ctx context.Context, token, userID string, ttl time.Duration) error {
	return c.client.Set(ctx, "challenge_token:"+token, userID, ttl).Err()
}

// GetUserByToken - look up user_id by challenge token (returns "" if not found/expired)
func (c *ChallengeCache) GetUserByToken(ctx context.Context, token string) (string, error) {
	val, err := c.client.Get(ctx, "challenge_token:"+token).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// DeleteToken - invalidate a used challenge token
func (c *ChallengeCache) DeleteToken(ctx context.Context, token string) error {
	return c.client.Del(ctx, "challenge_token:"+token).Err()
}
