package http

import "time"

type AuditLogResponse struct {
	ID            string         `json:"id"`
	AggregateType string         `json:"aggregate_type"`
	AggregateID   string         `json:"aggregate_id"`
	Action        string         `json:"action"`
	ActorID       string         `json:"actor_id"`
	Attributes    map[string]any `json:"attributes"`
	IPAddress     string         `json:"ip_address"`
	UserAgent     string         `json:"user_agent"`
	CreatedAt     time.Time      `json:"created_at"`
}

type EndpointHistoryResponse struct {
	ID         string    `json:"id"`
	Method     string    `json:"method"`
	Path       string    `json:"path"`
	StatusCode int       `json:"status_code"`
	UserID     string    `json:"user_id"`
	IPAddress  string    `json:"ip_address"`
	DurationMs int       `json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`
}
