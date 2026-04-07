package domain

import (
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// ErrorCode is a central registry entry mapping machine-readable codes
// to human-readable messages with i18n support.
type ErrorCode struct {
	domain.AggregateRoot
	Code       string // immutable after creation, e.g. "INSUFFICIENT_FUNDS"
	MessageEn  string
	MessageUz  string
	MessageRu  string
	Category   string // VALIDATION, SECURITY, DATA, BUSINESS, SYSTEM
	Severity   string // CRITICAL, HIGH, MEDIUM, LOW, INFO
	HTTPStatus int    // 400, 401, 403, 404, 409, 500
	Retryable  bool
	Suggestion string // end-user guidance
}

func NewErrorCode(code, msgEn, msgUz, msgRu, category, severity string, httpStatus int, retryable bool, suggestion string) (*ErrorCode, error) {
	if code == "" {
		return nil, ErrEmptyCode
	}
	now := time.Now()
	e := &ErrorCode{
		Code: code, MessageEn: msgEn, MessageUz: msgUz, MessageRu: msgRu,
		Category: category, Severity: severity, HTTPStatus: httpStatus,
		Retryable: retryable, Suggestion: suggestion,
	}
	e.CreatedAt = now
	e.UpdatedAt = now
	return e, nil
}

// Update modifies error code metadata (code itself is immutable).
func (e *ErrorCode) Update(msgEn, msgUz, msgRu, suggestion *string, httpStatus *int, retryable *bool) {
	if msgEn != nil {
		e.MessageEn = *msgEn
	}
	if msgUz != nil {
		e.MessageUz = *msgUz
	}
	if msgRu != nil {
		e.MessageRu = *msgRu
	}
	if suggestion != nil {
		e.Suggestion = *suggestion
	}
	if httpStatus != nil {
		e.HTTPStatus = *httpStatus
	}
	if retryable != nil {
		e.Retryable = *retryable
	}
	e.Touch()
}
