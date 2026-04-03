package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RecoveryMiddleware prevents the server from crashing on panic.
func RecoveryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				logger.Errorc(c.UserContext(), "panic recovered",
					zap.Any("panic", r),
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
					zap.String("stack", string(debug.Stack())),
				)

				apperror.ErrorHandler(c, apperror.ErrInternal.
					WithDetails(fmt.Sprintf("panic: %v", r)).
					WithSeverity(apperror.SeverityCritical))
			}
		}()

		return c.Next()
	}
}
