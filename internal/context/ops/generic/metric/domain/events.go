package domain

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

// MetricCollected is emitted when a new metric data point is recorded.
type MetricCollected struct {
	domain.BaseEvent
	Name  string
	Value float64
}

func NewMetricCollected(id, name string, value float64) MetricCollected {
	return MetricCollected{
		BaseEvent: domain.NewBaseEvent("metric.collected", id),
		Name:      name,
		Value:     value,
	}
}
