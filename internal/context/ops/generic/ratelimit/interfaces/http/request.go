package http

import "time"

type CreateRateLimitRequest struct {
	Key           string `json:"key"`
	MaxRequests   int    `json:"max_requests"`
	WindowSeconds int    `json:"window_seconds"`
	Description   string `json:"description"`
	Enabled       bool   `json:"enabled"`
}

type UpdateRateLimitRequest struct {
	MaxRequests   int    `json:"max_requests"`
	WindowSeconds int    `json:"window_seconds"`
	Description   string `json:"description"`
	Enabled       bool   `json:"enabled"`
}

type RateLimitResponse struct {
	ID            string    `json:"id"`
	Key           string    `json:"key"`
	MaxRequests   int       `json:"max_requests"`
	WindowSeconds int       `json:"window_seconds"`
	Description   string    `json:"description"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
