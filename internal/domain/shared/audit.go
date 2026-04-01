package shared

import (
	"context"
	"time"
)

// AuditEntry - a single audit log record
type AuditEntry struct {
	AggregateType string            // "Account", "Transfer", "Card"
	AggregateID   string            // entity ID
	Action        string            // "AccountOpened", "Credited", "TransferCompleted"
	ActorID       string            // user who performed the action
	Attributes    map[string]string // key-value event data
	IPAddress     string
	UserAgent     string
	Timestamp     time.Time
}

// AuditLog - interface for audit logging (MongoDB, etc.)
// Async - should not block the main flow
type AuditLog interface {
	Log(ctx context.Context, entry AuditEntry) error
}
