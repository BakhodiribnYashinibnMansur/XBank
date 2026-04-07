package middleware

import "github.com/gofiber/fiber/v2"

// HelmetMiddleware - sets security headers to protect against common web attacks
//
// Headers:
//   X-Content-Type-Options: nosniff       — prevents MIME type sniffing
//   X-Frame-Options: DENY                 — prevents clickjacking (iframe embedding)
//   X-XSS-Protection: 1; mode=block       — enables browser XSS filter
//   Strict-Transport-Security: max-age=.. — forces HTTPS for 1 year
//   Content-Security-Policy: default-src.. — restricts resource loading
//   Referrer-Policy: strict-origin...      — controls referrer header
//   Permissions-Policy: ...                — disables browser features (camera, mic, etc.)
//   Cache-Control: no-store                — prevents caching of sensitive responses
func HelmetMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		c.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
		c.Set("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Set("Pragma", "no-cache")

		return c.Next()
	}
}
