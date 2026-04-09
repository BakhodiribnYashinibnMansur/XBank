package http

import domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/metric/domain"

func toResponse(m *domain.AppMetric) MetricResponse {
	labels := m.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	return MetricResponse{
		ID:          m.ID,
		Name:        m.Name,
		Value:       m.Value,
		Labels:      labels,
		CollectedAt: m.CollectedAt,
	}
}
