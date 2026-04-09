package audit

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"go.uber.org/zap"
)

// Logger is the interface for security audit logging.
type Logger interface {
	Log(ctx context.Context, event Event) error
}

// ZapAuditLogger writes security events as structured JSON via a dedicated zap logger.
type ZapAuditLogger struct {
	logger *zap.Logger
}

// NewZapAuditLogger creates a logger backed by the given zap.Logger instance.
func NewZapAuditLogger(logger *zap.Logger) *ZapAuditLogger {
	return &ZapAuditLogger{logger: logger}
}

// Log writes a security event. If the event has no ID, one is generated.
func (l *ZapAuditLogger) Log(ctx context.Context, event Event) error {
	if event.ID == "" {
		event.ID = generateEventID()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	fields := []zap.Field{
		zap.String("event_id", event.ID),
		zap.String("type", string(event.Type)),
		zap.String("user_id", event.UserID),
		zap.String("ip_address", event.IPAddress),
		zap.String("user_agent", event.UserAgent),
		zap.Any("metadata", event.Metadata),
		zap.Time("timestamp", event.Timestamp),
	}

	l.logger.Info("security_audit", fields...)
	return nil
}

func generateEventID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
