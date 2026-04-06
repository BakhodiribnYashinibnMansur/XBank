package entity

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// File is an immutable aggregate root for file metadata.
// Actual binary is stored in MinIO; this tracks metadata only.
type File struct {
	domain.AggregateRoot
	Name         string // system-generated name (UUID-based)
	OriginalName string // user's original filename
	MimeType     string
	Size         int64  // bytes
	Path         string // MinIO path: "bucket/object"
	URL          string // download URL
	UploadedBy   string // user ID (empty = system upload)
}

func NewFile(name, originalName, mimeType string, size int64, path, url, uploadedBy string) (*File, error) {
	if name == "" || originalName == "" {
		return nil, ErrEmptyName
	}
	now := time.Now()
	f := &File{
		Name: name, OriginalName: originalName, MimeType: mimeType,
		Size: size, Path: path, URL: url, UploadedBy: uploadedBy,
	}
	f.CreatedAt = now
	f.UpdatedAt = now
	return f, nil
}
