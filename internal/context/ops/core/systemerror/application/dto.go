package application

type RecordErrorRequest struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Severity   string            `json:"severity"`
	Category   string            `json:"category"`
	StackTrace string            `json:"stack_trace"`
	RequestID  string            `json:"request_id"`
	UserID     string            `json:"user_id"`
	IPAddress  string            `json:"ip_address"`
	Path       string            `json:"path"`
	Method     string            `json:"method"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}
