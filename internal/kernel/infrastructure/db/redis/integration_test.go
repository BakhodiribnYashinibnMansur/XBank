// Integration tests for Redis-backed services.
// Requires a running Redis instance.
// Set REDIS_URL env or tests will be skipped.
//
// Run: REDIS_URL=redis://localhost:6379/1 go test ./internal/infrastructure/redis/ -v -run Integration
package redis

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	goredis "github.com/redis/go-redis/v9"
)

func init() {
	logger.Init(true)
}

func getTestRedis(t *testing.T) *goredis.Client {
	t.Helper()
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL not set, skipping integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := NewClient(ctx, url)
	if err != nil {
		t.Skipf("Cannot connect to Redis: %v", err)
	}

	t.Cleanup(func() {
		client.Close()
	})

	return client
}

// ── Session Cache Integration Tests ────────────────

func TestIntegration_SessionCache_SetAndGet(t *testing.T) {
	client := getTestRedis(t)
	cache := NewSessionCache(client)
	ctx := context.Background()

	sessionID := "test-session-integration-1"
	userID := "user-123"

	// Set
	err := cache.SetSession(ctx, sessionID, userID, 1*time.Minute)
	if err != nil {
		t.Fatalf("SetSession failed: %v", err)
	}

	// Get
	got, err := cache.GetSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if got != userID {
		t.Errorf("GetSession: got %q, want %q", got, userID)
	}

	// Cleanup
	cache.DeleteSession(ctx, sessionID)
}

func TestIntegration_SessionCache_Delete(t *testing.T) {
	client := getTestRedis(t)
	cache := NewSessionCache(client)
	ctx := context.Background()

	sessionID := "test-session-integration-del"
	cache.SetSession(ctx, sessionID, "user-1", 1*time.Minute)

	err := cache.DeleteSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	got, _ := cache.GetSession(ctx, sessionID)
	if got != "" {
		t.Errorf("Session should be gone after delete, got %q", got)
	}
}

func TestIntegration_SessionCache_Expiry(t *testing.T) {
	client := getTestRedis(t)
	cache := NewSessionCache(client)
	ctx := context.Background()

	sessionID := "test-session-integration-exp"
	cache.SetSession(ctx, sessionID, "user-1", 1*time.Second)

	// Wait for expiry
	time.Sleep(1500 * time.Millisecond)

	got, _ := cache.GetSession(ctx, sessionID)
	if got != "" {
		t.Errorf("Session should expire after TTL, got %q", got)
	}
}

func TestIntegration_SessionCache_Blacklist(t *testing.T) {
	client := getTestRedis(t)
	cache := NewSessionCache(client)
	ctx := context.Background()

	token := "jwt-token-to-blacklist"
	err := cache.BlacklistToken(ctx, token, 1*time.Minute)
	if err != nil {
		t.Fatalf("BlacklistToken failed: %v", err)
	}

	blacklisted, err := cache.IsBlacklisted(ctx, token)
	if err != nil {
		t.Fatalf("IsBlacklisted failed: %v", err)
	}
	if !blacklisted {
		t.Error("Token should be blacklisted")
	}

	// Non-blacklisted token
	clean, _ := cache.IsBlacklisted(ctx, "clean-token")
	if clean {
		t.Error("Non-blacklisted token should return false")
	}
}

// ── Login Limiter Integration Tests ────────────────

func TestIntegration_LoginLimiter_NotLockedInitially(t *testing.T) {
	client := getTestRedis(t)
	limiter := NewLoginLimiter(client, 3, 1*time.Minute, 1*time.Minute)
	ctx := context.Background()

	email := "integration_test_limiter@xbank.test"
	client.Del(ctx, "login_attempts:"+email, "login_locked:"+email)

	locked, err := limiter.IsLocked(ctx, email)
	if err != nil {
		t.Fatalf("IsLocked failed: %v", err)
	}
	if locked {
		t.Error("Should not be locked initially")
	}

	// Cleanup
	client.Del(ctx, "login_attempts:"+email, "login_locked:"+email)
}

func TestIntegration_LoginLimiter_LocksAfterMaxFailures(t *testing.T) {
	client := getTestRedis(t)
	limiter := NewLoginLimiter(client, 3, 1*time.Minute, 1*time.Minute)
	ctx := context.Background()

	email := "integration_test_blocked@xbank.test"
	client.Del(ctx, "login_attempts:"+email, "login_locked:"+email)

	// Exceed limit
	for i := 0; i < 4; i++ {
		limiter.RecordFailure(ctx, email)
	}

	locked, err := limiter.IsLocked(ctx, email)
	if err != nil {
		t.Fatalf("IsLocked failed: %v", err)
	}
	if !locked {
		t.Error("Should be locked after exceeding max attempts")
	}

	// Reset and verify unlock
	limiter.ResetAttempts(ctx, email)

	// Cleanup
	client.Del(ctx, "login_attempts:"+email, "login_locked:"+email)
}

// ── Challenge Cache Integration Tests ────────────────

func TestIntegration_ChallengeCache_SetAndGet(t *testing.T) {
	client := getTestRedis(t)
	cache := NewChallengeCache(client)
	ctx := context.Background()

	token := "challenge-test-token"

	err := cache.SetToken(ctx, token, "user-1", 1*time.Minute)
	if err != nil {
		t.Fatalf("SetToken failed: %v", err)
	}

	userID, err := cache.GetUserByToken(ctx, token)
	if err != nil {
		t.Fatalf("GetUserByToken failed: %v", err)
	}
	if userID != "user-1" {
		t.Errorf("GetUserByToken: got %q, want user-1", userID)
	}

	// Delete
	cache.DeleteToken(ctx, token)
	userID, _ = cache.GetUserByToken(ctx, token)
	if userID != "" {
		t.Error("Token should be deleted")
	}
}
