package cookie

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

// Options configures cookie attributes.
type Options struct {
	MaxAge   time.Duration
	Path     string
	Domain   string
	Secure   bool
	HTTPOnly bool
	SameSite string // "Lax", "Strict", "None"
}

// DefaultOptions returns secure defaults suitable for production.
func DefaultOptions(maxAge time.Duration) Options {
	return Options{
		MaxAge:   maxAge,
		Path:     "/",
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Lax",
	}
}

// Set writes a cookie to the response.
func Set(c *fiber.Ctx, name, value string, opts Options) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     opts.Path,
		Domain:   opts.Domain,
		MaxAge:   int(opts.MaxAge.Seconds()),
		Secure:   opts.Secure,
		HTTPOnly: opts.HTTPOnly,
		SameSite: opts.SameSite,
	})
}

// Get reads a cookie value from the request. Returns empty string if not found.
func Get(c *fiber.Ctx, name string) string {
	return c.Cookies(name)
}

// Delete removes a cookie by setting MaxAge to -1.
func Delete(c *fiber.Ctx, name string) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Lax",
	})
}

// SetRefreshToken writes the refresh token cookie with secure defaults.
func SetRefreshToken(c *fiber.Ctx, token string, ttl time.Duration) {
	Set(c, "refresh_token", token, Options{
		MaxAge:   ttl,
		Path:     "/api/v1/auth",
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

// DeleteRefreshToken removes the refresh token cookie.
func DeleteRefreshToken(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}
