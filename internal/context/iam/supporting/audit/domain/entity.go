package domain

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// Domain errors
var (
	ErrAuditLogNotFound = domain.NewDomainError("AUDIT_LOG_NOT_FOUND", "audit log not found")
	ErrInvalidAggregate = domain.NewDomainError("INVALID_AGGREGATE", "aggregate_type and aggregate_id are required")
	ErrInvalidAction    = domain.NewDomainError("INVALID_ACTION", "action cannot be empty")
)

// AuditLog represents a recorded audit event.
type AuditLog struct {
	ID            string
	AggregateType string
	AggregateID   string
	Action        string
	ActorID       string
	Attributes    map[string]any
	IPAddress     string
	UserAgent     string
	CreatedAt     time.Time
}

// NewAuditLog creates and validates an AuditLog.
func NewAuditLog(aggregateType, aggregateID, action, actorID string, attrs map[string]any, ip, ua string) (*AuditLog, error) {
	if aggregateType == "" || aggregateID == "" {
		return nil, ErrInvalidAggregate
	}
	if action == "" {
		return nil, ErrInvalidAction
	}
	if attrs == nil {
		attrs = map[string]any{}
	}
	return &AuditLog{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Action:        action,
		ActorID:       actorID,
		Attributes:    attrs,
		IPAddress:     ip,
		UserAgent:     ua,
		CreatedAt:     time.Now(),
	}, nil
}

// EndpointHistory tracks API endpoint access.
type EndpointHistory struct {
	ID         string
	Method     string
	Path       string
	StatusCode int
	UserID     string
	IPAddress  string
	DurationMs int
	CreatedAt  time.Time
}

// NewEndpointHistory creates and validates an EndpointHistory entry.
func NewEndpointHistory(method, path string, statusCode int, userID, ip string, durationMs int) (*EndpointHistory, error) {
	return &EndpointHistory{
		Method:     method,
		Path:       path,
		StatusCode: statusCode,
		UserID:     userID,
		IPAddress:  ip,
		DurationMs: durationMs,
		CreatedAt:  time.Now(),
	}, nil
}

// AuditFilter defines query parameters for listing audit logs.
type AuditFilter struct {
	AggregateType string
	AggregateID   string
	Action        string
	ActorID       string
	From          *time.Time
	To            *time.Time
	Limit         int
	Offset        int
}

// EndpointFilter defines query parameters for listing endpoint history.
type EndpointFilter struct {
	Method string
	Path   string
	UserID string
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}

// Repository defines the persistence contract for audit entities.
type Repository interface {
	CreateAuditLog(ctx context.Context, log *AuditLog) error
	ListAuditLogs(ctx context.Context, filter AuditFilter) ([]*AuditLog, int64, error)

	CreateEndpointHistory(ctx context.Context, h *EndpointHistory) error
	ListEndpointHistory(ctx context.Context, filter EndpointFilter) ([]*EndpointHistory, int64, error)
}
