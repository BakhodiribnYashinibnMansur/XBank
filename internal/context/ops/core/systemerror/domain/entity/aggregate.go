package entity

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

type ResolutionStatus string

const (
	StatusPending  ResolutionStatus = "PENDING"
	StatusResolved ResolutionStatus = "RESOLVED"
)

// SystemError tracks 5xx errors with full request context for debugging.
// Resolution is one-way: PENDING → RESOLVED (irreversible).
type SystemError struct {
	domain.AggregateRoot
	Code       string
	Message    string
	Severity   string // CRITICAL, HIGH, MEDIUM, LOW
	Category   string // SYSTEM, DATA, EXTERNAL, NETWORK
	StackTrace string
	RequestID  string
	UserID     string
	IPAddress  string
	Path       string
	Method     string
	Metadata   map[string]string
	Resolution ResolutionStatus
	ResolvedAt *time.Time
	ResolvedBy string
}

// NewSystemError creates a new error record.
func NewSystemError(code, message, severity, category string) (*SystemError, error) {
	if code == "" {
		return nil, ErrEmptyCode
	}
	now := time.Now()
	e := &SystemError{
		Code:       code,
		Message:    message,
		Severity:   severity,
		Category:   category,
		Resolution: StatusPending,
	}
	e.CreatedAt = now
	e.UpdatedAt = now
	return e, nil
}

// WithContext enriches the error with request context.
func (e *SystemError) WithContext(requestID, userID, ipAddress, path, method, stackTrace string, metadata map[string]string) {
	e.RequestID = requestID
	e.UserID = userID
	e.IPAddress = ipAddress
	e.Path = path
	e.Method = method
	e.StackTrace = stackTrace
	e.Metadata = metadata
}

// Resolve marks the error as resolved. One-way transition.
func (e *SystemError) Resolve(resolvedBy string) error {
	if e.Resolution == StatusResolved {
		return ErrAlreadyResolved
	}
	now := time.Now()
	e.Resolution = StatusResolved
	e.ResolvedAt = &now
	e.ResolvedBy = resolvedBy
	e.Touch()
	return nil
}
