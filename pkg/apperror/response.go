package apperror

import (
	"errors"
	"net/http"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// APIResponse is the unified envelope for ALL API responses (success + error).
//
// Success example:
//
//	{"status":"success","code":200,"data":{"id":"..."},"meta":{"request_id":"...","timestamp":"..."}}
//
// Error example:
//
//	{"status":"error","code":3004,"message":"Invalid email or password","data":null,"meta":{"request_id":"...","timestamp":"..."}}
type APIResponse struct {
	Status  string      `json:"status"`            // "success" or "error"
	Code    int         `json:"code"`              // numeric: 200, 201 for success; 1001-9999 for errors
	Message string      `json:"message,omitempty"` // human-readable error message (empty on success)
	Data    interface{} `json:"data"`              // response payload (null on error)
	Meta    Meta        `json:"meta"`
}

// Meta contains request-level metadata for tracing.
type Meta struct {
	RequestID     string `json:"request_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
	Timestamp     string `json:"timestamp"`
}

// --- Keep old types for backward compatibility with Swagger annotations ---

// ErrorBody is used in Swagger docs.
type ErrorBody struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse is used in Swagger docs.
type ErrorResponse struct {
	Status string    `json:"status"`
	Error  ErrorBody `json:"error"`
	Meta   Meta      `json:"meta"`
}

// SuccessResponse is used in Swagger docs.
type SuccessResponse struct {
	Status string      `json:"status"`
	Code   int         `json:"code"`
	Data   interface{} `json:"data"`
	Meta   Meta        `json:"meta,omitempty"`
}

// --- Unified response builders ---

func buildMeta(c *fiber.Ctx) Meta {
	requestID, _ := c.Locals("request_id").(string)
	correlationID := c.Get("X-Correlation-ID")
	return Meta{
		RequestID:     requestID,
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

// Success creates a standard success response.
//
//	return apperror.Success(c, http.StatusOK, data)
//	→ {"status":"success","code":200,"data":{...},"meta":{...}}
func Success(c *fiber.Ctx, httpStatus int, data interface{}) error {
	return c.Status(httpStatus).JSON(APIResponse{
		Status: "success",
		Code:   httpStatus,
		Data:   data,
		Meta:   buildMeta(c),
	})
}

// ErrorHandler is a Fiber error handler that converts errors to unified JSON.
//
// AppError:
//
//	→ {"status":"error","code":3004,"message":"Invalid email or password","data":null,"meta":{...}}
//
// Fiber error (404, etc.):
//
//	→ {"status":"error","code":404,"message":"Not Found","data":null,"meta":{...}}
//
// Unknown error:
//
//	→ {"status":"error","code":1000,"message":"Internal server error","data":null,"meta":{...}}
func ErrorHandler(c *fiber.Ctx, err error) error {
	meta := buildMeta(c)
	ctx := c.UserContext()

	var appErr *AppError
	if errors.As(err, &appErr) {
		logger.LogAppError(ctx, appErr)
		return c.Status(appErr.HTTPStatus).JSON(APIResponse{
			Status:  "error",
			Code:    appErr.Code,
			Message: appErr.Message,
			Data:    nil,
			Meta:    meta,
		})
	}

	// DomainError → map to AppError automatically
	if IsDomainError(err) {
		mapped := MapDomainError(err)
		logger.LogAppError(ctx, mapped)
		return c.Status(mapped.HTTPStatus).JSON(APIResponse{
			Status:  "error",
			Code:    mapped.Code,
			Message: mapped.Message,
			Data:    nil,
			Meta:    meta,
		})
	}

	// Fiber errors (404 route not found, etc.)
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		logger.Warnc(ctx, "fiber_error",
			zap.Int("status", fiberErr.Code),
			zap.String("message", fiberErr.Message),
		)
		return c.Status(fiberErr.Code).JSON(APIResponse{
			Status:  "error",
			Code:    fiberErr.Code,
			Message: fiberErr.Message,
			Data:    nil,
			Meta:    meta,
		})
	}

	// Unknown error
	logger.Errorc(ctx, "unhandled_error", zap.Error(err))
	return c.Status(http.StatusInternalServerError).JSON(APIResponse{
		Status:  "error",
		Code:    ErrInternal.Code,
		Message: ErrInternal.Message,
		Data:    nil,
		Meta:    meta,
	})
}
