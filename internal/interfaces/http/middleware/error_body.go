package middleware

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ErrorBodyMiddleware logs the request body when the response status is 4xx/5xx.
// Helps debug client errors and server failures without enabling full body logging.
func ErrorBodyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Capture body before handler consumes it
		body := c.Body()

		err := c.Next()

		status := c.Response().StatusCode()
		if status >= 400 && len(body) > 0 {
			// Truncate large bodies
			logBody := string(body)
			if len(logBody) > 2048 {
				logBody = logBody[:2048] + "...[truncated]"
			}

			logger.Log.Warn("error response with request body",
				zap.Int("status", status),
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.String("body", logBody),
			)
		}

		return err
	}
}
