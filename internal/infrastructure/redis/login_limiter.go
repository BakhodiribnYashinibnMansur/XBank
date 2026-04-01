package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// LoginLimiter - Redis-based brute-force protection for login
//
// After maxAttempts failed logins within the window:
//   - Account is locked for lockDuration
//   - Even correct password is rejected during lock
type LoginLimiter struct {
	client       *goredis.Client
	maxAttempts  int
	window       time.Duration
	lockDuration time.Duration
}

func NewLoginLimiter(client *goredis.Client, maxAttempts int, window, lockDuration time.Duration) *LoginLimiter {
	return &LoginLimiter{
		client:       client,
		maxAttempts:  maxAttempts,
		window:       window,
		lockDuration: lockDuration,
	}
}

// IsLocked - check if the account is locked due to failed attempts
func (l *LoginLimiter) IsLocked(ctx context.Context, email string) (bool, error) {
	if l.client == nil {
		return false, nil
	}
	exists, err := l.client.Exists(ctx, l.lockKey(email)).Result()
	return exists > 0, err
}

// RecordFailure - increment failed login counter
// Returns true if account is now locked
func (l *LoginLimiter) RecordFailure(ctx context.Context, email string) (bool, error) {
	if l.client == nil {
		return false, nil
	}

	key := l.attemptKey(email)

	count, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// Set TTL on first attempt
	if count == 1 {
		l.client.Expire(ctx, key, l.window)
	}

	// Lock if max attempts reached
	if int(count) >= l.maxAttempts {
		l.client.Set(ctx, l.lockKey(email), "1", l.lockDuration)
		l.client.Del(ctx, key) // reset counter
		return true, nil
	}

	return false, nil
}

// ResetAttempts - clear failed attempts on successful login
func (l *LoginLimiter) ResetAttempts(ctx context.Context, email string) {
	if l.client == nil {
		return
	}
	l.client.Del(ctx, l.attemptKey(email))
}

func (l *LoginLimiter) attemptKey(email string) string {
	return fmt.Sprintf("login_attempts:%s", email)
}

func (l *LoginLimiter) lockKey(email string) string {
	return fmt.Sprintf("login_locked:%s", email)
}
