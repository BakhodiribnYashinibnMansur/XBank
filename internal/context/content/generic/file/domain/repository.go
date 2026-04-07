package domain

import "context"

type WriteRepository interface {
	Save(ctx context.Context, f interface{}) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (interface{}, error)
}

type FileView struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	OriginalName string `json:"original_name"`
	MimeType     string `json:"mime_type"`
	Size         int64  `json:"size"`
	URL          string `json:"url"`
	UploadedBy   string `json:"uploaded_by"`
	CreatedAt    string `json:"created_at"`
}

type FileFilter struct {
	MimeType string
	Limit    int
	Offset   int
}

type ReadRepository interface {
	FindByID(ctx context.Context, id string) (*FileView, error)
	List(ctx context.Context, filter FileFilter) ([]*FileView, int64, error)
}
