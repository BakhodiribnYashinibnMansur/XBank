package handler

import (
	authApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/auth"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *authApp.Service
}

func NewAuthHandler(authService *authApp.Service) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login godoc
// @Summary      User login
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.LoginRequest true "Login credentials"
// @Success      200 {object} dto.AuthResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	if req.Email == "" || req.Password == "" {
		return apperror.ErrMissingField.WithMessage("email and password are required")
	}

	result, err := h.authService.Login(c.Context(), req.Email, req.Password, c.Get("User-Agent"), c.IP())
	if err != nil {
		return err
	}

	return c.JSON(dto.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User: dto.UserResponse{
			ID:        result.User.ID,
			Email:     result.User.Email,
			FirstName: result.User.FirstName,
			LastName:  result.User.LastName,
			CreatedAt: result.User.CreatedAt,
		},
	})
}

// Refresh godoc
// @Summary      Refresh access token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.RefreshRequest true "Refresh token"
// @Success      200 {object} dto.AuthResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req dto.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	if req.RefreshToken == "" {
		return apperror.ErrMissingField.WithMessage("refresh_token is required")
	}

	result, err := h.authService.Refresh(c.Context(), req.RefreshToken, c.Get("User-Agent"), c.IP())
	if err != nil {
		return err
	}

	return c.JSON(dto.AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User: dto.UserResponse{
			ID:        result.User.ID,
			Email:     result.User.Email,
			FirstName: result.User.FirstName,
			LastName:  result.User.LastName,
			CreatedAt: result.User.CreatedAt,
		},
	})
}

// Logout godoc
// @Summary      Logout (end session)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.LogoutRequest true "Refresh token"
// @Success      200 {object} map[string]string
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	var req dto.LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	h.authService.Logout(c.Context(), req.RefreshToken)

	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}

// LogoutAll godoc
// @Summary      Logout from all sessions
// @Tags         Auth
// @Produce      json
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.authService.LogoutAll(c.Context(), userID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "All sessions terminated"})
}
