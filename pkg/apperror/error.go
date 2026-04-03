package apperror

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// AppError is the structured error type used across all layers.
type AppError struct {
	Code       int            `json:"code"`
	Message    string         `json:"message"`
	HTTPStatus int            `json:"-"`
	Err        error          `json:"-"`

	// Type is a machine-readable error discriminator (e.g. "USER_NOT_FOUND").
	Type string `json:"-"`
	// UserMsg is a user-facing message (safe to display in UI).
	UserMsg string `json:"user_message,omitempty"`
	// Severity indicates how critical the error is (CRITICAL, HIGH, MEDIUM, LOW, INFO).
	Severity ErrorSeverity `json:"-"`
	// Category classifies the error domain (VALIDATION, SECURITY, DATA, BUSINESS, SYSTEM, EXTERNAL).
	Category ErrorCategory `json:"-"`
	// Source identifies where the error originated (e.g. "UserRepo.GetByID").
	Source string `json:"-"`
	// Details provides additional developer-facing context.
	Details string `json:"-"`
	// Fields holds arbitrary key-value metadata for logging.
	Fields map[string]any `json:"-"`
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

// copy creates a shallow clone so sentinel values stay immutable.
func (e *AppError) copy() *AppError {
	cp := *e
	if e.Fields != nil {
		cp.Fields = make(map[string]any, len(e.Fields))
		for k, v := range e.Fields {
			cp.Fields[k] = v
		}
	}
	return &cp
}

func (e *AppError) Wrap(err error) *AppError {
	cp := e.copy()
	cp.Err = err
	return cp
}

func (e *AppError) WithMessage(msg string) *AppError {
	cp := e.copy()
	cp.Message = msg
	return cp
}

func (e *AppError) WithType(t string) *AppError {
	cp := e.copy()
	cp.Type = t
	return cp
}

func (e *AppError) WithUserMsg(msg string) *AppError {
	cp := e.copy()
	cp.UserMsg = msg
	return cp
}

func (e *AppError) WithSeverity(s ErrorSeverity) *AppError {
	cp := e.copy()
	cp.Severity = s
	return cp
}

func (e *AppError) WithCategory(c ErrorCategory) *AppError {
	cp := e.copy()
	cp.Category = c
	return cp
}

func (e *AppError) WithSource(src string) *AppError {
	cp := e.copy()
	cp.Source = src
	return cp
}

func (e *AppError) WithDetails(d string) *AppError {
	cp := e.copy()
	cp.Details = d
	return cp
}

func (e *AppError) WithField(key string, val any) *AppError {
	cp := e.copy()
	if cp.Fields == nil {
		cp.Fields = make(map[string]any)
	}
	cp.Fields[key] = val
	return cp
}

// --- Logger integration (satisfies logger.LoggableError) ---

// LogFields returns structured fields for logging without import cycles.
func (e *AppError) LogFields() []zap.Field {
	fields := []zap.Field{
		zap.String("error_type", e.Type),
		zap.Int("error_code", e.Code),
		zap.Int("http_status", e.HTTPStatus),
		zap.String("message", e.Message),
		zap.String("severity", string(e.Severity)),
		zap.String("category", string(e.Category)),
	}
	if e.Source != "" {
		fields = append(fields, zap.String("source", e.Source))
	}
	if e.Details != "" {
		fields = append(fields, zap.String("details", e.Details))
	}
	if e.Err != nil {
		fields = append(fields, zap.Error(e.Err))
	}
	for k, v := range e.Fields {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}

// LogLevel returns the appropriate log level based on severity.
func (e *AppError) LogLevel() string {
	switch e.Severity {
	case SeverityCritical, SeverityHigh:
		return "error"
	case SeverityMedium:
		return "warn"
	default:
		return "info"
	}
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
