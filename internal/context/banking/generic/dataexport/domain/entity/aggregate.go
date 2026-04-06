package entity

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

type ExportStatus string

const (
	ExportPending    ExportStatus = "PENDING"
	ExportProcessing ExportStatus = "PROCESSING"
	ExportCompleted  ExportStatus = "COMPLETED"
	ExportFailed     ExportStatus = "FAILED"
)

// DataExport tracks a GDPR data export request lifecycle.
// State machine: PENDING → PROCESSING → COMPLETED | FAILED
type DataExport struct {
	domain.AggregateRoot
	UserID   string
	Status   ExportStatus
	FileURL  string // set on completion
	ErrorMsg string // set on failure
}

func NewDataExport(userID string) (*DataExport, error) {
	if userID == "" {
		return nil, ErrEmptyUserID
	}
	now := time.Now()
	e := &DataExport{UserID: userID, Status: ExportPending}
	e.CreatedAt = now
	e.UpdatedAt = now
	return e, nil
}

func (e *DataExport) StartProcessing() error {
	if e.Status != ExportPending {
		return ErrInvalidTransition
	}
	e.Status = ExportProcessing
	e.Touch()
	return nil
}

func (e *DataExport) Complete(fileURL string) error {
	if e.Status != ExportProcessing {
		return ErrInvalidTransition
	}
	e.Status = ExportCompleted
	e.FileURL = fileURL
	e.Touch()
	return nil
}

func (e *DataExport) Fail(reason string) error {
	if e.Status != ExportProcessing {
		return ErrInvalidTransition
	}
	e.Status = ExportFailed
	e.ErrorMsg = reason
	e.Touch()
	return nil
}
