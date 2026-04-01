package middleware

import (
	"strconv"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/gofiber/fiber/v2"
)

// MetricsMiddleware records HTTP request metrics:
//   - xbank_http_requests_total (counter) — by method, path, status
//   - xbank_http_request_duration_seconds (histogram) — by method, path
//   - xbank_http_active_requests (gauge) — in-flight count
func MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		metrics.HTTPActiveRequests.Inc()
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())
		method := c.Method()
		path := c.Route().Path // normalized route pattern, not raw URL

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
		metrics.HTTPActiveRequests.Dec()

		return err
	}
}
