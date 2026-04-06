package application

// CreateTranslationRequest DTO.
type CreateTranslationRequest struct {
	Key      string `json:"key"`
	Language string `json:"language"`
	Value    string `json:"value"`
	Group    string `json:"group"`
}

// UpdateTranslationRequest DTO.
type UpdateTranslationRequest struct {
	Value string `json:"value"`
}
