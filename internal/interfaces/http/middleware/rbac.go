package middleware

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// RequireRole - checks that the user has one of the allowed roles
// Usage: router.Post("/admin/action", middleware.RequireRole("ADMIN"), handler)
func RequireRole(allowedRoles ...string) fiber.Handler {
	roleSet := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = true
	}

	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		if !roleSet[role] {
			return apperror.ErrForbidden
		}
		return c.Next()
	}
}
