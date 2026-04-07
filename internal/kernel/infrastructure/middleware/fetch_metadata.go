package middleware

import (
	"crypto/sha256"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// FetchMetadataMiddleware extracts IP, User-Agent, and a device fingerprint
// from request headers and stores them in Locals for downstream handlers.
// Keys: "client_ip", "user_agent", "device_fingerprint"
func FetchMetadataMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		ua := c.Get("User-Agent")
		acceptLang := c.Get("Accept-Language")

		// Deterministic fingerprint from stable headers
		raw := fmt.Sprintf("%s|%s|%s", ua, acceptLang, c.Get("Sec-CH-UA"))
		hash := sha256.Sum256([]byte(raw))
		fingerprint := fmt.Sprintf("%x", hash[:16]) // 32-char hex

		c.Locals("client_ip", ip)
		c.Locals("user_agent", ua)
		c.Locals("device_fingerprint", fingerprint)

		return c.Next()
	}
}
