package http

// CreateFlagRequest is the HTTP request for creating a feature flag.
type CreateFlagRequest struct {
	Key          string `json:"key"`
	Description  string `json:"description"`
	FlagType     string `json:"flag_type"`     // bool, string, int, float
	DefaultValue string `json:"default_value"`
}

// UpdateFlagRequest is the HTTP request for updating a feature flag.
type UpdateFlagRequest struct {
	Description  *string `json:"description,omitempty"`
	DefaultValue *string `json:"default_value,omitempty"`
	Active       *bool   `json:"active,omitempty"`
	RolloutPct   *int    `json:"rollout_pct,omitempty"`
}

// EvaluateFlagRequest is the HTTP request for evaluating a flag.
type EvaluateFlagRequest struct {
	Key        string            `json:"key"`
	UserID     string            `json:"user_id"`
	Attributes map[string]string `json:"attributes,omitempty"`
}
