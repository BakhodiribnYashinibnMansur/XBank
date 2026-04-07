package middleware

import (
	"os"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// DebugBodyMiddleware logs full request and response bodies in development mode.
// Only active when APP_ENV=development. Never enable in production.
func DebugBodyMiddleware() fiber.Handler {
	enabled := os.Getenv("APP_ENV") == "development"

	return func(c *fiber.Ctx) error {
		if !enabled {
			return c.Next()
		}

		reqBody := string(c.Body())
		if len(reqBody) > 4096 {
			reqBody = reqBody[:4096] + "...[truncated]"
		}

		logger.Log.Debug("request body",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("body", reqBody),
		)

		err := c.Next()

		respBody := string(c.Response().Body())
		if len(respBody) > 4096 {
			respBody = respBody[:4096] + "...[truncated]"
		}

		logger.Log.Debug("response body",
			zap.Int("status", c.Response().StatusCode()),
			zap.String("body", respBody),
		)

		return err
	}
}
