package domain

import "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"

type ErrorRecorded struct {
	domain.BaseEvent
	Code     string
	Severity string
}

func NewErrorRecorded(id, code, severity string) ErrorRecorded {
	return ErrorRecorded{
		BaseEvent: domain.NewBaseEvent("systemerror.recorded", id),
		Code:      code,
		Severity:  severity,
	}
}

type ErrorResolved struct {
	domain.BaseEvent
	ResolvedBy string
}

func NewErrorResolved(id, resolvedBy string) ErrorResolved {
	return ErrorResolved{
		BaseEvent:  domain.NewBaseEvent("systemerror.resolved", id),
		ResolvedBy: resolvedBy,
	}
}
