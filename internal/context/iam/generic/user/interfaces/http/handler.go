package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/user/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *command.Service
}

func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	if req.Email == "" || req.Password == "" || req.FirstName == "" {
		return apperror.ErrMissingField.WithMessage("email, password and first_name are required")
	}

	if len(req.Password) < 8 {
		return apperror.ErrInvalidPassword
	}

	u, err := h.service.Register(c.Context(), req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
	})
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id query parameter is required")
	}

	u, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
	})
}

func (h *Handler) ChangePassword(c *fiber.Ctx) error {
	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		return apperror.ErrMissingField.WithMessage("old_password and new_password are required")
	}

	userID := c.Locals("user_id").(string)
	if err := h.service.ChangePassword(c.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Password changed"})
}

func (h *Handler) ExportData(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	u, err := h.service.ExportData(c.Context(), userID)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, UserResponse{
		ID: u.ID, Email: u.Email, FirstName: u.FirstName, LastName: u.LastName, CreatedAt: u.CreatedAt,
	})
}

func (h *Handler) DeleteAccount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.service.DeleteAccount(c.Context(), userID); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Account data anonymized"})
}
