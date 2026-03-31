package middleware

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RecoveryMiddleware - prevents the server from crashing on panic
func RecoveryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				logger.Log.Error("panic caught",
					zap.Any("error", r),
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
				)

				apperror.ErrorHandler(c, apperror.ErrInternal)
			}
		}()

		return c.Next()
	}
}
