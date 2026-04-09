package domain

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

var (
	ErrCheckNotFound = domain.NewDomainError("CHECK_NOT_FOUND", "health check not found")
)

// ComponentStatus represents the health status of a component.
type ComponentStatus string

const (
	StatusHealthy   ComponentStatus = "HEALTHY"
	StatusDegraded  ComponentStatus = "DEGRADED"
	StatusUnhealthy ComponentStatus = "UNHEALTHY"
)

// ComponentCheck represents the health check result for a single component.
type ComponentCheck struct {
	Name      string
	Status    ComponentStatus
	Latency   time.Duration // response time of the check
	Message   string        // optional details or error message
	CheckedAt time.Time
}

// SystemHealth represents the overall system health state.
type SystemHealth struct {
	Status     ComponentStatus
	Components []ComponentCheck
	CheckedAt  time.Time
}

// NewSystemHealth creates a SystemHealth from individual component checks.
// Overall status is the worst status among all components.
func NewSystemHealth(checks []ComponentCheck) *SystemHealth {
	overall := StatusHealthy
	for _, c := range checks {
		if c.Status == StatusUnhealthy {
			overall = StatusUnhealthy
			break
		}
		if c.Status == StatusDegraded && overall == StatusHealthy {
			overall = StatusDegraded
		}
	}
	return &SystemHealth{
		Status:     overall,
		Components: checks,
		CheckedAt:  time.Now(),
	}
}

// HealthRecord represents a persisted health check snapshot.
type HealthRecord struct {
	ID         string
	Status     ComponentStatus
	Components string // JSON-serialized component checks
	CheckedAt  time.Time
}

// HealthChecker defines an interface for checking component health.
type HealthChecker interface {
	Check(ctx context.Context) ComponentCheck
}

// Repository defines the persistence interface for health records.
type Repository interface {
	Save(ctx context.Context, record *HealthRecord) error
	GetLatest(ctx context.Context) (*HealthRecord, error)
	ListHistory(ctx context.Context, limit, offset int) ([]*HealthRecord, error)
	CountHistory(ctx context.Context) (int64, error)
}
