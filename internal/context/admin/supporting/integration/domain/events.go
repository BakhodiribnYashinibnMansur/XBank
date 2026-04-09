package domain

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

// IntegrationCreated is emitted when a new integration is registered.
type IntegrationCreated struct {
	domain.BaseEvent
	Name   string
	Status string
}

func NewIntegrationCreated(id, name string, status Status) IntegrationCreated {
	return IntegrationCreated{
		BaseEvent: domain.NewBaseEvent("integration.created", id),
		Name:      name,
		Status:    string(status),
	}
}

// IntegrationDeleted is emitted when an integration is removed.
type IntegrationDeleted struct {
	domain.BaseEvent
}

func NewIntegrationDeleted(id string) IntegrationDeleted {
	return IntegrationDeleted{
		BaseEvent: domain.NewBaseEvent("integration.deleted", id),
	}
}
