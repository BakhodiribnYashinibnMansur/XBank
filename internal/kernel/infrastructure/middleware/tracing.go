package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "xbank.http.middleware"

// fiberCarrier adapts Fiber request headers to the OpenTelemetry TextMapCarrier interface.
type fiberCarrier struct {
	ctx *fiber.Ctx
}

func (c fiberCarrier) Get(key string) string {
	return c.ctx.Get(key)
}

func (c fiberCarrier) Set(key, value string) {
	c.ctx.Set(key, value)
}

func (c fiberCarrier) Keys() []string {
	keys := make([]string, 0)
	c.ctx.Request().Header.VisitAll(func(k, v []byte) {
		keys = append(keys, string(k))
	})
	return keys
}

// TracingMiddleware creates a span for each HTTP request and propagates trace context.
//
// Span attributes:
//   - http.method, http.route, http.status_code, http.url
//   - net.host.name, user_agent, client.address
//
// The trace_id and span_id are stored in c.Locals for use by LoggerMiddleware.
func TracingMiddleware() fiber.Handler {
	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	return func(c *fiber.Ctx) error {
		// Extract parent span context from incoming headers (W3C Trace Context)
		carrier := fiberCarrier{ctx: c}
		ctx := propagator.Extract(c.Context(), carrier)

		// Start a new span
		spanName := fmt.Sprintf("%s %s", c.Method(), c.Path())
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(c.Method()),
				semconv.URLPath(c.Path()),
				semconv.URLFull(c.OriginalURL()),
				semconv.ServerAddress(c.Hostname()),
				semconv.UserAgentOriginal(c.Get("User-Agent")),
				semconv.ClientAddress(c.IP()),
			),
		)
		defer span.End()

		// Store trace context in Fiber's UserContext so downstream code can use it
		c.SetUserContext(ctx)

		// Store trace_id and span_id in Locals for logger middleware
		sc := span.SpanContext()
		if sc.HasTraceID() {
			c.Locals("trace_id", sc.TraceID().String())
		}
		if sc.HasSpanID() {
			c.Locals("span_id", sc.SpanID().String())
		}

		// Inject trace context into response headers (for downstream services)
		propagator.Inject(ctx, carrier)

		// Execute the next handler
		err := c.Next()

		// Record response status
		status := c.Response().StatusCode()
		span.SetAttributes(semconv.HTTPResponseStatusCode(status))

		if status >= 500 {
			span.SetAttributes(attribute.Bool("error", true))
		}

		if err != nil {
			span.RecordError(err)
		}

		return err
	}
}
