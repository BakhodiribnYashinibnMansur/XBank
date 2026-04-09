package middleware

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
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
			userID, _ := c.Locals("user_id").(string)
			logger.Log.Warn("authz_failure: insufficient role",
				zap.String("user_id", userID),
				zap.String("role", role),
				zap.Strings("required_roles", allowedRoles),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.String("ip", c.IP()),
			)
			return apperror.ErrForbidden
		}
		return c.Next()
	}
}

// AccessCheckFunc is a function that checks RBAC access.
// Returns (allowed, scope, error). Scope is "own" or "all".
type AccessCheckFunc func(ctx context.Context, roleName, resource, action string) (allowed bool, scope string, err error)

// RequireAccess returns middleware that checks fine-grained RBAC policies.
// Usage: router.Post("/accounts", middleware.RequireAccess(checkFn, "accounts", "write"), handler)
func RequireAccess(check AccessCheckFunc, resource, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		allowed, scope, err := check(c.Context(), role, resource, action)
		if err != nil || !allowed {
			userID, _ := c.Locals("user_id").(string)
			logger.Log.Warn("authz_failure: access denied",
				zap.String("user_id", userID),
				zap.String("role", role),
				zap.String("resource", resource),
				zap.String("action", action),
				zap.String("path", c.Path()),
				zap.String("ip", c.IP()),
			)
			return apperror.ErrForbidden
		}
		c.Locals("authz_scope", scope)
		return c.Next()
	}
}
