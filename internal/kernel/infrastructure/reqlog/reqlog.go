package reqlog

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/redact"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Fields extracts structured log fields from a Fiber request context.
// Sensitive headers (Authorization, Cookie) are automatically redacted.
func Fields(c *fiber.Ctx) []zap.Field {
	fields := []zap.Field{
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.String("ip", c.IP()),
		zap.String("user_agent", c.Get("User-Agent")),
	}

	if reqID, ok := c.Locals("request_id").(string); ok && reqID != "" {
		fields = append(fields, zap.String("request_id", reqID))
	}
	if corrID := c.Get("X-Correlation-ID"); corrID != "" {
		fields = append(fields, zap.String("correlation_id", corrID))
	}
	if auth := c.Get("Authorization"); auth != "" {
		fields = append(fields, zap.String("authorization", redact.Bearer(auth)))
	}

	return fields
}

// Response adds response-related fields (status, duration, size).
func Response(c *fiber.Ctx, status int, start time.Time) []zap.Field {
	return []zap.Field{
		zap.Int("status", status),
		zap.Duration("duration", time.Since(start)),
		zap.Int("body_size", len(c.Response().Body())),
	}
}

// Entry logs a complete request/response cycle at the appropriate level.
func Entry(l *zap.Logger, c *fiber.Ctx, status int, start time.Time) {
	fields := append(Fields(c), Response(c, status, start)...)

	switch {
	case status >= 500:
		l.Error("http_request", fields...)
	case status >= 400:
		l.Warn("http_request", fields...)
	default:
		l.Info("http_request", fields...)
	}
}
