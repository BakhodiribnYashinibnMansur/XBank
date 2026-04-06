package http

type CreateNotificationRequest struct {
	UserID  string            `json:"user_id"`
	Title   string            `json:"title"`
	Message string            `json:"message"`
	Type    string            `json:"type"`
	Data    map[string]string `json:"data,omitempty"`
}
