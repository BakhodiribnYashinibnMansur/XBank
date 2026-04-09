package http

import domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/integration/domain"

func toResponse(i *domain.Integration) IntegrationResponse {
	return IntegrationResponse{
		ID:         i.ID,
		Name:       i.Name,
		BaseURL:    i.BaseURL,
		APIKey:     i.APIKey,
		Status:     string(i.Status),
		WebhookURL: i.WebhookURL,
		CreatedAt:  i.CreatedAt,
		UpdatedAt:  i.UpdatedAt,
	}
}
