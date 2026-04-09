package http

import "time"

type MetricResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Value       float64           `json:"value"`
	Labels      map[string]string `json:"labels"`
	CollectedAt time.Time         `json:"collected_at"`
}
