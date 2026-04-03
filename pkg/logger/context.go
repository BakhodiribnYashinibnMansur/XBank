package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxLoggerKey struct{}

// WithLogger stores a child logger in the context.
func WithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, l)
}

// FromContext extracts the logger from context, falling back to the global Log.
func FromContext(ctx context.Context) *zap.Logger {
	if ctx != nil {
		if l, ok := ctx.Value(ctxLoggerKey{}).(*zap.Logger); ok {
			return l
		}
	}
	return Log
}

// Debugc logs at Debug level using the context-scoped logger.
func Debugc(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Debug(msg, fields...)
}

// Infoc logs at Info level using the context-scoped logger.
func Infoc(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Info(msg, fields...)
}

// Warnc logs at Warn level using the context-scoped logger.
func Warnc(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Warn(msg, fields...)
}

// Errorc logs at Error level using the context-scoped logger.
func Errorc(ctx context.Context, msg string, fields ...zap.Field) {
	FromContext(ctx).Error(msg, fields...)
}
