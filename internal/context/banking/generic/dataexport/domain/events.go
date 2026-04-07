package domain

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

type ExportRequested struct {
	domain.BaseEvent
	UserID string
}

func NewExportRequested(id, userID string) ExportRequested {
	return ExportRequested{BaseEvent: domain.NewBaseEvent("dataexport.requested", id), UserID: userID}
}

type ExportCompleted struct {
	domain.BaseEvent
	UserID  string
	FileURL string
}

func NewExportCompleted(id, userID, fileURL string) ExportCompleted {
	return ExportCompleted{BaseEvent: domain.NewBaseEvent("dataexport.completed", id), UserID: userID, FileURL: fileURL}
}
