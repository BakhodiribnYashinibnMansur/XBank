package handler

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

	return c.Status(http.StatusCreated).JSON(dto.UserResponse{
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

	return c.JSON(dto.UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
	})
}
