package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
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
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Juda ko'p so'rov. Biroz kutib turing.",
			})
		},
	})
}
