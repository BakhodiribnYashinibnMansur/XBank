package middleware

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/gofiber/fiber/v2"
)

// RLSMiddleware - injects user_id into context for PostgreSQL Row-Level Security
// Must be placed AFTER AuthMiddleware (needs user_id from JWT)
func RLSMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, _ := c.Locals("user_id").(string)
		if userID != "" {
			ctx := postgres.WithRLSUser(c.Context(), userID)
			c.SetUserContext(ctx)
		}
		return c.Next()
	}
}
