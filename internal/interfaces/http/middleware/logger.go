package middleware

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// LoggerMiddleware - logs every request using zap
//
// Development: 2026-03-31T12:00:00  INFO  http.request  {"method":"GET","path":"/health","status":200,"duration":"0.3ms"}
// Production:  {"level":"info","ts":1711872000,"msg":"http.request","method":"GET","path":"/health","status":200,"duration":"0.3ms"}
func LoggerMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		requestID, _ := c.Locals("request_id").(string)
		traceID, _ := c.Locals("trace_id").(string)
		spanID, _ := c.Locals("span_id").(string)

		logger.Log.Info("http.request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", time.Since(start)),
			zap.String("ip", c.IP()),
			zap.String("request_id", requestID),
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
		)

		return err
	}
}
