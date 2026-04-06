package http

type CreateTranslationRequest struct {
	Key      string `json:"key"`
	Language string `json:"language"`
	Value    string `json:"value"`
	Group    string `json:"group"`
}

type UpdateTranslationRequest struct {
	Value string `json:"value"`
}
