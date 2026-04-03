package logger

import (
	"context"

	"go.uber.org/zap"
)

// LoggableError is satisfied by apperror.AppError without importing it (no cycle).
type LoggableError interface {
	error
	LogFields() []zap.Field
	LogLevel() string // "error", "warn", "info"
}

// LogAppError logs a LoggableError with full metadata at the appropriate level.
func LogAppError(ctx context.Context, err LoggableError) {
	l := FromContext(ctx)
	fields := err.LogFields()

	switch err.LogLevel() {
	case "error":
		l.Error("app_error", fields...)
	case "warn":
		l.Warn("app_error", fields...)
	default:
		l.Info("app_error", fields...)
	}
}
