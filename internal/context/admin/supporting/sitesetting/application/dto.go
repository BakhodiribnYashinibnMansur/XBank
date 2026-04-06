package application

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/sitesetting/domain/entity"

// CreateSettingRequest is the DTO for creating a site setting.
type CreateSettingRequest struct {
	Key         string             `json:"key"`
	Value       string             `json:"value"`
	SettingType entity.SettingType `json:"setting_type"`
	Description string             `json:"description"`
}

// UpdateSettingRequest is the DTO for updating a site setting.
type UpdateSettingRequest struct {
	Value       *string `json:"value,omitempty"`
	Description *string `json:"description,omitempty"`
}

// SettingResponse is the DTO returned to clients.
type SettingResponse struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	SettingType string `json:"setting_type"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
