package response

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Envelope is the standard API response envelope.
type Envelope struct {
	Status  string `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data"`
	Meta    Meta   `json:"meta"`
}

// Meta holds request-level metadata for tracing.
type Meta struct {
	RequestID     string `json:"request_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
	Timestamp     string `json:"timestamp"`
}

func buildMeta(c *fiber.Ctx) Meta {
	requestID, _ := c.Locals("request_id").(string)
	correlationID := c.Get("X-Correlation-ID")
	return Meta{
		RequestID:     requestID,
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

// OK sends a 200 success response.
func OK(c *fiber.Ctx, data any) error {
	return Success(c, http.StatusOK, data)
}

// Created sends a 201 success response.
func Created(c *fiber.Ctx, data any) error {
	return Success(c, http.StatusCreated, data)
}

// NoContent sends a 204 response with no body.
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}

// Success sends a success response with the given HTTP status and data.
func Success(c *fiber.Ctx, httpStatus int, data any) error {
	return c.Status(httpStatus).JSON(Envelope{
		Status: "success",
		Code:   httpStatus,
		Data:   data,
		Meta:   buildMeta(c),
	})
}

// Error sends an error response with the given HTTP status, app code, and message.
func Error(c *fiber.Ctx, httpStatus int, code int, message string) error {
	return c.Status(httpStatus).JSON(Envelope{
		Status:  "error",
		Code:    code,
		Message: message,
		Data:    nil,
		Meta:    buildMeta(c),
	})
}

// Paginated sends a paginated success response.
func Paginated(c *fiber.Ctx, data any, total int64, limit, offset int) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "success",
		"code":   http.StatusOK,
		"data":   data,
		"pagination": fiber.Map{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
		"meta": buildMeta(c),
	})
}
