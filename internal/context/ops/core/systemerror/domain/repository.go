package domain

import "context"

type WriteRepository interface {
	Save(ctx context.Context, e interface{}) error
	Update(ctx context.Context, e interface{}) error
	FindByID(ctx context.Context, id string) (interface{}, error)
}

type SystemErrorView struct {
	ID         string            `json:"id"`
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Severity   string            `json:"severity"`
	Category   string            `json:"category"`
	StackTrace string            `json:"stack_trace,omitempty"`
	RequestID  string            `json:"request_id"`
	UserID     string            `json:"user_id,omitempty"`
	IPAddress  string            `json:"ip_address"`
	Path       string            `json:"path"`
	Method     string            `json:"method"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Resolution string            `json:"resolution"`
	ResolvedBy string            `json:"resolved_by,omitempty"`
	CreatedAt  string            `json:"created_at"`
}

type SystemErrorFilter struct {
	Code       string
	Severity   string
	Resolution string
	DateFrom   string
	DateTo     string
	Limit      int
	Offset     int
}

type ReadRepository interface {
	FindByID(ctx context.Context, id string) (*SystemErrorView, error)
	List(ctx context.Context, filter SystemErrorFilter) ([]*SystemErrorView, int64, error)
}
