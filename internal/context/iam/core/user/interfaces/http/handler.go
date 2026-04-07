package http

import (
	"net/http"

	userApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/user"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	service *userApp.Service
}

func NewUserHandler(service *userApp.Service) *UserHandler {
	return &UserHandler{service: service}
}

// Register godoc
// @Summary      Register a new user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.RegisterRequest true "Registration data"
// @Success      201 {object} dto.UserResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      409 {object} apperror.ErrorResponse
// @Router       /auth/register [post]
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
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

	return apperror.Success(c, http.StatusCreated, dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
	})
}

// GetByID godoc
// @Summary      Get user by ID
// @Tags         Users
// @Produce      json
// @Param        id query string true "User ID"
// @Success      200 {object} dto.UserResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /users/get [get]
func (h *UserHandler) GetByID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id query parameter is required")
	}

	u, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
	})
}

// ChangePassword godoc
// @Summary      Change password
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body body dto.ChangePasswordRequest true "Old and new password"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /users/change-password [post]
func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	var req dto.ChangePasswordRequest
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

// ExportData godoc
// @Summary      Export user data (GDPR)
// @Tags         Users
// @Produce      json
// @Success      200 {object} dto.UserResponse
// @Security     BearerAuth
// @Router       /users/me/data-export [get]
func (h *UserHandler) ExportData(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	u, err := h.service.ExportData(c.Context(), userID)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, dto.UserResponse{
		ID: u.ID, Email: u.Email, FirstName: u.FirstName, LastName: u.LastName, CreatedAt: u.CreatedAt,
	})
}

// DeleteAccount godoc
// @Summary      Delete account / anonymize data (GDPR)
// @Tags         Users
// @Produce      json
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /users/me/delete [delete]
func (h *UserHandler) DeleteAccount(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	if err := h.service.DeleteAccount(c.Context(), userID); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Account data anonymized"})
}
