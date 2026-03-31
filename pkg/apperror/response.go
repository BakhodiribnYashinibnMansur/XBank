package apperror

import (
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ErrorBody is the structured error returned in JSON responses.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is the top-level error response envelope.
type ErrorResponse struct {
	Status string    `json:"status"`
	Error  ErrorBody `json:"error"`
	Meta   Meta      `json:"meta"`
}

// Meta contains request-level metadata for tracing.
type Meta struct {
	RequestID     string `json:"request_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
	Timestamp     string `json:"timestamp"`
}

// ErrorHandler is a Fiber error handler that converts AppError to structured JSON.
// Set as app.Config.ErrorHandler.
func ErrorHandler(c *fiber.Ctx, err error) error {
	requestID, _ := c.Locals("requestID").(string)
	correlationID := c.Get("X-Correlation-ID")

	meta := Meta{
		RequestID:     requestID,
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.HTTPStatus).JSON(ErrorResponse{
			Status: "error",
			Error: ErrorBody{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
			Meta: meta,
		})
	}

	// Fiber errors (404 route not found, etc.)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return c.Status(fiberErr.Code).JSON(ErrorResponse{
			Status: "error",
			Error: ErrorBody{
				Code:    http.StatusText(fiberErr.Code),
				Message: fiberErr.Message,
			},
			Meta: meta,
		})
	}

	// Unknown error — always 500 with generic message
	return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
		Status: "error",
		Error: ErrorBody{
			Code:    ErrInternal.Code,
			Message: ErrInternal.Message,
		},
		Meta: meta,
	})
}
