package domain

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

type ErrorCodeCreated struct {
	domain.BaseEvent
	Code string
}

func NewErrorCodeCreated(id, code string) ErrorCodeCreated {
	return ErrorCodeCreated{BaseEvent: domain.NewBaseEvent("errorcode.created", id), Code: code}
}

type ErrorCodeUpdated struct {
	domain.BaseEvent
	Code string
}

func NewErrorCodeUpdated(id, code string) ErrorCodeUpdated {
	return ErrorCodeUpdated{BaseEvent: domain.NewBaseEvent("errorcode.updated", id), Code: code}
}
