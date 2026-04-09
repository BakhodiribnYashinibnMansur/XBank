package errorx

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// Wrap wraps an error with a contextual message.
func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// Wrapf wraps an error with a formatted contextual message.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// IsAppError checks if the error chain contains an *apperror.AppError.
func IsAppError(err error) bool {
	var appErr *apperror.AppError
	return errors.As(err, &appErr)
}

// AsAppError extracts the *apperror.AppError from the error chain.
func AsAppError(err error) (*apperror.AppError, bool) {
	var appErr *apperror.AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}

// HTTPStatus extracts the HTTP status from an error.
// Returns 500 for unknown errors.
func HTTPStatus(err error) int {
	if appErr, ok := AsAppError(err); ok {
		return appErr.HTTPStatus
	}
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return fiberErr.Code
	}
	return http.StatusInternalServerError
}

// IsNotFound returns true if the error represents a 404.
func IsNotFound(err error) bool {
	return HTTPStatus(err) == http.StatusNotFound
}

// IsConflict returns true if the error represents a 409.
func IsConflict(err error) bool {
	return HTTPStatus(err) == http.StatusConflict
}

// IsUnauthorized returns true if the error represents a 401.
func IsUnauthorized(err error) bool {
	return HTTPStatus(err) == http.StatusUnauthorized
}
