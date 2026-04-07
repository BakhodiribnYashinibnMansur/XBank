package domain

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/generic/dataexport/domain/entity"
)

type WriteRepository interface {
	Save(ctx context.Context, e *entity.DataExport) error
	Update(ctx context.Context, e *entity.DataExport) error
	FindByID(ctx context.Context, id string) (*entity.DataExport, error)
}

type DataExportView struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Status    string `json:"status"`
	FileURL   string `json:"file_url,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
	CreatedAt string `json:"created_at"`
}

type DataExportFilter struct {
	UserID string
	Status string
	Limit  int
	Offset int
}

type ReadRepository interface {
	FindByID(ctx context.Context, id string) (*DataExportView, error)
	List(ctx context.Context, filter DataExportFilter) ([]*DataExportView, int64, error)
}
