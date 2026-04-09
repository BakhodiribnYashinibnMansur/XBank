package validator

import (
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// BodyParser parses and validates the request body into the target struct.
// Returns apperror.ErrInvalidJSON if parsing fails.
func BodyParser[T any](c *fiber.Ctx) (T, error) {
	var req T
	if err := c.BodyParser(&req); err != nil {
		return req, apperror.ErrInvalidJSON
	}
	return req, nil
}

// Validatable is implemented by request DTOs that can validate themselves.
type Validatable interface {
	Validate() error
}

// BodyParseAndValidate parses the request body and calls Validate() on it.
func BodyParseAndValidate[T Validatable](c *fiber.Ctx) (T, error) {
	req, err := BodyParser[T](c)
	if err != nil {
		return req, err
	}
	if err := req.Validate(); err != nil {
		return req, err
	}
	return req, nil
}

// QueryParam extracts a required query parameter. Returns apperror.ErrMissingField if empty.
func QueryParam(c *fiber.Ctx, name string) (string, error) {
	v := c.Query(name)
	if v == "" {
		return "", apperror.ErrMissingField.WithMessage(name + " is required")
	}
	return v, nil
}

// PathParam extracts a required path parameter. Returns apperror.ErrInvalidParam if empty.
func PathParam(c *fiber.Ctx, name string) (string, error) {
	v := c.Params(name)
	if v == "" {
		return "", apperror.ErrInvalidParam.WithMessage(name + " is required")
	}
	return v, nil
}
