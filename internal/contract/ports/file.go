package ports

import (
	"context"
	"io"
)

// FileStorage provides file upload/download capabilities for other BCs.
type FileStorage interface {
	Upload(ctx context.Context, req UploadRequest) (*UploadResult, error)
	GetURL(ctx context.Context, fileID string) (string, error)
	Delete(ctx context.Context, fileID string) error
}

// UploadRequest describes a file to be uploaded.
type UploadRequest struct {
	FileName    string
	ContentType string
	Size        int64
	Reader      io.Reader
	UploadedBy  string
}

// UploadResult contains the metadata of an uploaded file.
type UploadResult struct {
	FileID string `json:"file_id"`
	URL    string `json:"url"`
	Path   string `json:"path"`
}
