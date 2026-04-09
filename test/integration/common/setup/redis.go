package setup

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RedisContainer wraps a testcontainers Redis instance.
type RedisContainer struct {
	Container testcontainers.Container
	Client    *redis.Client
	Addr      string
}

// MustRedis starts a Redis container and returns a connected client.
// The container is cleaned up when the test finishes.
func MustRedis(t *testing.T) *RedisContainer {
	t.Helper()
	ctx := context.Background()

	container, err := tcredis.Run(ctx,
		"redis:7-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(15*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("starting redis container: %v", err)
	}
	t.Cleanup(func() { container.Terminate(ctx) })

	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("getting redis endpoint: %v", err)
	}

	client := redis.NewClient(&redis.Options{Addr: endpoint})
	t.Cleanup(func() { client.Close() })

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("pinging redis: %v", err)
	}

	return &RedisContainer{
		Container: container,
		Client:    client,
		Addr:      endpoint,
	}
}

// FlushRedis clears all data from Redis for test isolation.
func FlushRedis(t *testing.T, client *redis.Client) {
	t.Helper()
	if err := client.FlushAll(context.Background()).Err(); err != nil {
		t.Fatalf("flushing redis: %v", err)
	}
}
