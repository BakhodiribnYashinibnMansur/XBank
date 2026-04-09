package domain

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

var (
	ErrIntegrationNotFound = domain.NewDomainError("INTEGRATION_NOT_FOUND", "integration not found")
	ErrIntegrationExists   = domain.NewDomainError("INTEGRATION_EXISTS", "integration with this name already exists")
)

type Status string

const (
	StatusActive    Status = "ACTIVE"
	StatusInactive  Status = "INACTIVE"
	StatusSuspended Status = "SUSPENDED"
)

// Integration represents an external service integration.
type Integration struct {
	ID         string
	Name       string
	BaseURL    string
	APIKey     string
	Status     Status
	WebhookURL string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewIntegration creates a new integration.
func NewIntegration(name, baseURL, apiKey string, status Status, webhookURL string) (*Integration, error) {
	if name == "" {
		return nil, domain.NewDomainError("MISSING_FIELD", "name is required")
	}
	if baseURL == "" {
		return nil, domain.NewDomainError("MISSING_FIELD", "base_url is required")
	}
	if status != StatusActive && status != StatusInactive && status != StatusSuspended {
		return nil, domain.NewDomainError("INVALID_FIELD", "status must be ACTIVE, INACTIVE, or SUSPENDED")
	}

	now := time.Now()
	return &Integration{
		Name:       name,
		BaseURL:    baseURL,
		APIKey:     apiKey,
		Status:     status,
		WebhookURL: webhookURL,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// Update modifies the integration fields.
func (i *Integration) Update(baseURL, apiKey string, status Status, webhookURL string) {
	i.BaseURL = baseURL
	i.APIKey = apiKey
	i.Status = status
	i.WebhookURL = webhookURL
	i.UpdatedAt = time.Now()
}

// Repository defines the persistence contract for integrations.
type Repository interface {
	Save(ctx context.Context, i *Integration) error
	FindByID(ctx context.Context, id string) (*Integration, error)
	FindByName(ctx context.Context, name string) (*Integration, error)
	ListAll(ctx context.Context) ([]*Integration, error)
	Update(ctx context.Context, i *Integration) error
	Delete(ctx context.Context, id string) error
}
