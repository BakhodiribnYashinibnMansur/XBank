package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/utils"
)

// CSRFMiddleware - protects against Cross-Site Request Forgery attacks
//
// How it works:
//   1. Server generates a CSRF token and sets it in a cookie (csrf_token)
//   2. Client must send the token in X-CSRF-Token header with mutating requests
//   3. If token is missing or invalid → 403 Forbidden
//
// Exempt paths: login, register, refresh (public endpoints)
// Only applies to: POST, PUT, PATCH, DELETE
func CSRFMiddleware() fiber.Handler {
	return csrf.New(csrf.Config{
		KeyLookup:      "header:X-CSRF-Token",
		CookieName:     "csrf_token",
		CookieSameSite: "Strict",
		CookieSecure:   true,
		CookieHTTPOnly: true,
		Expiration:     1 * time.Hour,
		KeyGenerator:   utils.UUIDv4,
		Next: func(c *fiber.Ctx) bool {
			// Skip CSRF for these paths (public or API-only)
			path := c.Path()
			switch path {
			case "/api/v1/auth/login",
				"/api/v1/auth/register",
				"/api/v1/auth/refresh",
				"/api/v1/auth/logout",
				"/health",
				"/health/live",
				"/health/ready":
				return true
			}
			// Skip for GET, HEAD, OPTIONS (safe methods)
			method := c.Method()
			if method == "GET" || method == "HEAD" || method == "OPTIONS" {
				return true
			}
			return false
		},
	})
}
