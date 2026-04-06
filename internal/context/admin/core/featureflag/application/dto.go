package application

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain/entity"

// CreateFlagRequest is the DTO for creating a feature flag.
type CreateFlagRequest struct {
	Key          string          `json:"key"`
	Description  string          `json:"description"`
	FlagType     entity.FlagType `json:"flag_type"`
	DefaultValue string          `json:"default_value"`
}

// UpdateFlagRequest is the DTO for updating a feature flag.
type UpdateFlagRequest struct {
	Description  *string `json:"description,omitempty"`
	DefaultValue *string `json:"default_value,omitempty"`
	Active       *bool   `json:"active,omitempty"`
	RolloutPct   *int    `json:"rollout_pct,omitempty"`
}

// EvaluateRequest is the DTO for evaluating a flag.
type EvaluateRequest struct {
	Key        string            `json:"key"`
	UserID     string            `json:"user_id"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// EvaluateResponse is the result of flag evaluation.
type EvaluateResponse struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}
