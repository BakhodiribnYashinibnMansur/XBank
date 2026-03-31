package apperror

import (
	"fmt"
	"net/http"
)

// AppError is the structured error type used across all layers.
// It carries an error code, HTTP status, human-readable message,
// and an optional wrapped cause for debugging.
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Is allows errors.Is comparison by code.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// Wrap returns a copy of this error with a wrapped cause.
func (e *AppError) Wrap(err error) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		HTTPStatus: e.HTTPStatus,
		Err:        err,
	}
}

// WithMessage returns a copy with an overridden message.
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    msg,
		HTTPStatus: e.HTTPStatus,
		Err:        e.Err,
	}
}

// --- Constructors ---

func New(code string, httpStatus int, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// --- HTTP helpers ---

func BadRequest(code, message string) *AppError {
	return New(code, http.StatusBadRequest, message)
}

func Unauthorized(code, message string) *AppError {
	return New(code, http.StatusUnauthorized, message)
}

func Forbidden(code, message string) *AppError {
	return New(code, http.StatusForbidden, message)
}

func NotFound(code, message string) *AppError {
	return New(code, http.StatusNotFound, message)
}

func Conflict(code, message string) *AppError {
	return New(code, http.StatusConflict, message)
}

func TooManyRequests(code, message string) *AppError {
	return New(code, http.StatusTooManyRequests, message)
}

func Internal(code, message string) *AppError {
	return New(code, http.StatusInternalServerError, message)
}

func ServiceUnavailable(code, message string) *AppError {
	return New(code, http.StatusServiceUnavailable, message)
}
