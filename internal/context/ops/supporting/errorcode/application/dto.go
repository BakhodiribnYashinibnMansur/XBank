package application

type CreateErrorCodeRequest struct {
	Code       string `json:"code"`
	MessageEn  string `json:"message_en"`
	MessageUz  string `json:"message_uz"`
	MessageRu  string `json:"message_ru"`
	Category   string `json:"category"`
	Severity   string `json:"severity"`
	HTTPStatus int    `json:"http_status"`
	Retryable  bool   `json:"retryable"`
	Suggestion string `json:"suggestion"`
}

type UpdateErrorCodeRequest struct {
	MessageEn  *string `json:"message_en,omitempty"`
	MessageUz  *string `json:"message_uz,omitempty"`
	MessageRu  *string `json:"message_ru,omitempty"`
	HTTPStatus *int    `json:"http_status,omitempty"`
	Retryable  *bool   `json:"retryable,omitempty"`
	Suggestion *string `json:"suggestion,omitempty"`
}
