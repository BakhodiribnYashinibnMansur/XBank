package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/session/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *command.Service
}

func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	if req.Email == "" || req.Password == "" {
		return apperror.ErrMissingField.WithMessage("email and password are required")
	}

	result, err := h.service.Login(c.Context(), req.Email, req.Password, c.Get("User-Agent"), c.IP())
	if err != nil {
		return err
	}

	if result.TOTPRequired {
		return apperror.Success(c, http.StatusOK, AuthResponse{
			TOTPRequired: true,
			User: UserResponse{
				ID:    result.UserID,
				Email: result.Email,
			},
		})
	}

	return apperror.Success(c, http.StatusOK, AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User: UserResponse{
			ID:    result.UserID,
			Email: result.Email,
		},
	})
}

func (h *Handler) Refresh(c *fiber.Ctx) error {
	var req RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	if req.RefreshToken == "" {
		return apperror.ErrMissingField.WithMessage("refresh_token is required")
	}

	result, err := h.service.Refresh(c.Context(), req.RefreshToken, c.Get("User-Agent"), c.IP())
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User: UserResponse{
			ID:    result.UserID,
			Email: result.Email,
		},
	})
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	h.service.Logout(c.Context(), req.RefreshToken)

	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Logged out successfully"})
}

func (h *Handler) LogoutAll(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	if err := h.service.LogoutAll(c.Context(), userID); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "All sessions terminated"})
}

func (h *Handler) TOTPVerifyLogin(c *fiber.Ctx) error {
	var req TOTPVerifyLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.Email == "" || req.Code == "" {
		return apperror.ErrMissingField.WithMessage("email and code are required")
	}

	result, err := h.service.LoginWithTOTP(c.Context(), req.Email, req.Code, c.Get("User-Agent"), c.IP())
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, AuthResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User: UserResponse{
			ID:    result.UserID,
			Email: result.Email,
		},
	})
}

func (h *Handler) TOTPSetup(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	secret, url, err := h.service.SetupTOTP(c.Context(), userID)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, TOTPSetupResponse{
		Secret: secret,
		URL:    url,
	})
}

func (h *Handler) TOTPConfirmSetup(c *fiber.Ctx) error {
	var req TOTPVerifySetupRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.Code == "" {
		return apperror.ErrMissingField.WithMessage("code is required")
	}

	userID := c.Locals("user_id").(string)

	if err := h.service.VerifyAndEnableTOTP(c.Context(), userID, req.Code); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "2FA enabled successfully"})
}

func (h *Handler) TOTPDisable(c *fiber.Ctx) error {
	var req TOTPDisableRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.Password == "" {
		return apperror.ErrMissingField.WithMessage("password is required")
	}

	userID := c.Locals("user_id").(string)

	if err := h.service.DisableTOTP(c.Context(), userID, req.Password); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "2FA disabled"})
}
