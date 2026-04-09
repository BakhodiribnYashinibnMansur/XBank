package domain

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

var (
	ErrMetricNotFound = domain.NewDomainError("METRIC_NOT_FOUND", "metric not found")
)

// AppMetric represents a single metric data point.
type AppMetric struct {
	ID          string
	Name        string
	Value       float64
	Labels      map[string]string // stored as JSON
	CollectedAt time.Time
}

// NewAppMetric creates a new application metric.
func NewAppMetric(name string, value float64, labels map[string]string) (*AppMetric, error) {
	if name == "" {
		return nil, domain.NewDomainError("MISSING_FIELD", "name is required")
	}

	if labels == nil {
		labels = make(map[string]string)
	}

	return &AppMetric{
		Name:        name,
		Value:       value,
		Labels:      labels,
		CollectedAt: time.Now(),
	}, nil
}

// Repository defines the persistence contract for app metrics.
type Repository interface {
	Save(ctx context.Context, m *AppMetric) error
	FindByName(ctx context.Context, name string) ([]*AppMetric, error)
	ListRecent(ctx context.Context, limit int) ([]*AppMetric, error)
}
