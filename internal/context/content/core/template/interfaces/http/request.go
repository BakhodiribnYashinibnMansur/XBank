package http

import "time"

// CreateTemplateRequest represents the body for creating a template.
type CreateTemplateRequest struct {
	Slug    string `json:"slug"`
	Channel string `json:"channel"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	Locale  string `json:"locale"`
}

// UpdateBodyRequest represents the body for updating template content.
type UpdateBodyRequest struct {
	ID      string `json:"id"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// IDRequest represents a simple ID-only request body.
type IDRequest struct {
	ID string `json:"id"`
}

// TemplateResponse represents a template in API responses.
type TemplateResponse struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Channel   string    `json:"channel"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Locale    string    `json:"locale"`
	Status    string    `json:"status"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
