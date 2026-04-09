package logger

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/contextx"
	"go.uber.org/zap"
)

// New creates a new zap.Logger enriched with standard context fields.
func New(ctx context.Context, base *zap.Logger) *zap.Logger {
	fields := extractFields(ctx)
	if len(fields) == 0 {
		return base
	}
	return base.With(fields...)
}

// WithOperation returns a logger enriched with operation-specific fields.
func WithOperation(ctx context.Context, base *zap.Logger, operation string) *zap.Logger {
	return New(ctx, base).With(zap.String("operation", operation))
}

// WithComponent returns a logger enriched with a component tag.
func WithComponent(base *zap.Logger, component string) *zap.Logger {
	return base.With(zap.String("component", component))
}

// extractFields pulls standard metadata from context into zap fields.
func extractFields(ctx context.Context) []zap.Field {
	var fields []zap.Field
	if id := contextx.GetRequestID(ctx); id != "" {
		fields = append(fields, zap.String("request_id", id))
	}
	if id := contextx.GetCorrelationID(ctx); id != "" {
		fields = append(fields, zap.String("correlation_id", id))
	}
	if id := contextx.GetUserID(ctx); id != "" {
		fields = append(fields, zap.String("user_id", id))
	}
	if id := contextx.GetSessionID(ctx); id != "" {
		fields = append(fields, zap.String("session_id", id))
	}
	if id := contextx.GetTraceID(ctx); id != "" {
		fields = append(fields, zap.String("trace_id", id))
	}
	return fields
}
