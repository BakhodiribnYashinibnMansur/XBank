package http

import "time"

type CreateIntegrationRequest struct {
	Name       string `json:"name"`
	BaseURL    string `json:"base_url"`
	APIKey     string `json:"api_key"`
	Status     string `json:"status"` // ACTIVE, INACTIVE, SUSPENDED
	WebhookURL string `json:"webhook_url"`
}

type UpdateIntegrationRequest struct {
	BaseURL    string `json:"base_url"`
	APIKey     string `json:"api_key"`
	Status     string `json:"status"`
	WebhookURL string `json:"webhook_url"`
}

type IntegrationResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	BaseURL    string    `json:"base_url"`
	APIKey     string    `json:"api_key"`
	Status     string    `json:"status"`
	WebhookURL string    `json:"webhook_url,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
