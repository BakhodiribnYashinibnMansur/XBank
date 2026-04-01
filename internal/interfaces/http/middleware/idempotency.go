package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	goredis "github.com/redis/go-redis/v9"
)

// IdempotencyMiddleware - prevents duplicate POST requests
// Client sends X-Idempotency-Key header. If the same key is sent again
// within the TTL, the previous response is returned without re-executing.
func IdempotencyMiddleware(redisClient *goredis.Client, ttl time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only apply to POST/PUT/PATCH (mutating requests)
		if c.Method() != "POST" && c.Method() != "PUT" && c.Method() != "PATCH" {
			return c.Next()
		}

		key := c.Get("X-Idempotency-Key")
		if key == "" {
			return c.Next() // no key = no idempotency check
		}

		if redisClient == nil {
			return c.Next() // Redis unavailable = skip
		}

		redisKey := "idempotency:" + key
		ctx := context.Background()

		// Check if this key was already processed
		cached, err := redisClient.Get(ctx, redisKey).Bytes()
		if err == nil && len(cached) > 0 {
			// Return cached response
			c.Set("X-Idempotency-Cached", "true")
			c.Set("Content-Type", "application/json")
			return c.Send(cached)
		}

		// Execute the handler
		if err := c.Next(); err != nil {
			return err
		}

		// Cache the response body for the TTL
		body := c.Response().Body()
		if len(body) > 0 && c.Response().StatusCode() < 400 {
			redisClient.Set(ctx, redisKey, body, ttl)
		}

		return nil
	}
}
