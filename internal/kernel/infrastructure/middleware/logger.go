package middleware

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// LoggerMiddleware logs every request and injects a request-scoped logger into context.
//
// The child logger carries request_id, trace_id, and span_id so that every
// downstream call to logger.Infoc(ctx, ...) automatically includes them.
func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		requestID, _ := c.Locals("request_id").(string)
		traceID, _ := c.Locals("trace_id").(string)
		spanID, _ := c.Locals("span_id").(string)

		// Create a request-scoped child logger with pre-populated fields.
		reqLogger := logger.Log.With(
			zap.String("request_id", requestID),
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
		)

		// Inject into Fiber's UserContext so services can use logger.FromContext(ctx).
		ctx := logger.WithLogger(c.UserContext(), reqLogger)
		c.SetUserContext(ctx)

		err := c.Next()

		reqLogger.Info("http.request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", time.Since(start)),
			zap.String("ip", c.IP()),
		)

		return err
	}
}
