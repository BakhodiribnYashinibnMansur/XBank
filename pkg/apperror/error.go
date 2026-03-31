package apperror

import (
	"fmt"
	"net/http"
)

// AppError is the structured error type used across all layers.
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func (e *AppError) Wrap(err error) *AppError {
	return &AppError{Code: e.Code, Message: e.Message, HTTPStatus: e.HTTPStatus, Err: err}
}

func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{Code: e.Code, Message: msg, HTTPStatus: e.HTTPStatus, Err: e.Err}
}

// --- Constructors ---

func New(code, httpStatus int, message string) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

func BadRequest(code int, message string) *AppError {
	return New(code, http.StatusBadRequest, message)
}

func Unauthorized(code int, message string) *AppError {
	return New(code, http.StatusUnauthorized, message)
}

func Forbidden(code int, message string) *AppError {
	return New(code, http.StatusForbidden, message)
}

func NotFound(code int, message string) *AppError {
	return New(code, http.StatusNotFound, message)
}

func Conflict(code int, message string) *AppError {
	return New(code, http.StatusConflict, message)
}

func TooManyRequests(code int, message string) *AppError {
	return New(code, http.StatusTooManyRequests, message)
}

func Internal(code int, message string) *AppError {
	return New(code, http.StatusInternalServerError, message)
}

func ServiceUnavailable(code int, message string) *AppError {
	return New(code, http.StatusServiceUnavailable, message)
}
