package event

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

type FileUploaded struct {
	domain.BaseEvent
	OriginalName string
	UploadedBy   string
}

func NewFileUploaded(id, originalName, uploadedBy string) FileUploaded {
	return FileUploaded{
		BaseEvent:    domain.NewBaseEvent("file.uploaded", id),
		OriginalName: originalName,
		UploadedBy:   uploadedBy,
	}
}
