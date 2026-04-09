package http

import "time"

// ComponentCheckResponse represents a single component check result.
type ComponentCheckResponse struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	LatencyMs int64     `json:"latency_ms"`
	Message   string    `json:"message,omitempty"`
	CheckedAt time.Time `json:"checked_at"`
}

// SystemHealthResponse represents the overall system health.
type SystemHealthResponse struct {
	Status     string                   `json:"status"`
	Components []ComponentCheckResponse `json:"components"`
	CheckedAt  time.Time                `json:"checked_at"`
}

// HealthRecordResponse represents a persisted health check record.
type HealthRecordResponse struct {
	ID         string                   `json:"id"`
	Status     string                   `json:"status"`
	Components []ComponentCheckResponse `json:"components"`
	CheckedAt  time.Time                `json:"checked_at"`
}
