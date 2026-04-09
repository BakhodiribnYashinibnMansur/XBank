package middleware

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"go.uber.org/zap"
)

// RateLimitMiddleware - blocks too many requests from a single IP
//
// Why is this needed?
// 1. DDoS protection: a bot cannot send 1000 requests/second
// 2. Brute-force protection: prevents login password guessing
// 3. Resource protection: prevents server overload
func RateLimitMiddleware(maxRequests int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			logger.Log.Warn("rate_limit_hit",
				zap.String("ip", c.IP()),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
			)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Juda ko'p so'rov. Biroz kutib turing.",
			})
		},
	})
}

// UserRateLimitMiddleware - per-user rate limiting for authenticated routes.
// Uses user_id from JWT context as key, falling back to IP for unauthenticated requests.
// This prevents a single user from abusing the API even across multiple IPs.
func UserRateLimitMiddleware(maxRequests int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			userID, ok := c.Locals("user_id").(string)
			if ok && userID != "" {
				return "user:" + userID
			}
			return "ip:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			userID, _ := c.Locals("user_id").(string)
			logger.Log.Warn("user_rate_limit_hit",
				zap.String("user_id", userID),
				zap.String("ip", c.IP()),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
			)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Juda ko'p so'rov. Biroz kutib turing.",
			})
		},
	})
}
