package handler

import (
	"net/http"

	authApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/auth"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// TOTPHandler - handles TOTP 2FA setup, verification, and login
type TOTPHandler struct {
	authService *authApp.Service
}

func NewTOTPHandler(authService *authApp.Service) *TOTPHandler {
	return &TOTPHandler{authService: authService}
}

// VerifyLogin godoc
// @Summary      Complete login with TOTP code
// @Description  Called after login returns totp_required=true
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.TOTPVerifyLoginRequest true "Email and TOTP code"
// @Success      200 {object} dto.AuthResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Router       /auth/totp/verify [post]
func (h *TOTPHandler) VerifyLogin(c *fiber.Ctx) error {
	var req dto.TOTPVerifyLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.Email == "" || req.Code == "" {
		return apperror.ErrMissingField.WithMessage("email and code are required")
	}

	result, err := h.authService.LoginWithTOTP(c.Context(), req.Email, req.Code, c.Get("User-Agent"), c.IP())
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, dto.AuthResponse{
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

// Setup godoc
// @Summary      Generate TOTP secret and QR code URL
// @Description  Returns a secret and otpauth:// URL for scanning with authenticator app
// @Tags         Auth
// @Produce      json
// @Success      200 {object} dto.TOTPSetupResponse
// @Security     BearerAuth
// @Router       /auth/totp/setup [post]
func (h *TOTPHandler) Setup(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	secret, url, err := h.authService.SetupTOTP(c.Context(), userID)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, dto.TOTPSetupResponse{
		Secret: secret,
		URL:    url,
	})
}

// ConfirmSetup godoc
// @Summary      Verify TOTP code and enable 2FA
// @Description  User enters the 6-digit code from authenticator app to confirm setup
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.TOTPVerifySetupRequest true "6-digit TOTP code"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /auth/totp/confirm [post]
func (h *TOTPHandler) ConfirmSetup(c *fiber.Ctx) error {
	var req dto.TOTPVerifySetupRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.Code == "" {
		return apperror.ErrMissingField.WithMessage("code is required")
	}

	userID := c.Locals("user_id").(string)

	if err := h.authService.VerifyAndEnableTOTP(c.Context(), userID, req.Code); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "2FA enabled successfully"})
}

// Disable godoc
// @Summary      Disable 2FA
// @Description  Requires password confirmation
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.TOTPDisableRequest true "Password"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /auth/totp/disable [post]
func (h *TOTPHandler) Disable(c *fiber.Ctx) error {
	var req dto.TOTPDisableRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.Password == "" {
		return apperror.ErrMissingField.WithMessage("password is required")
	}

	userID := c.Locals("user_id").(string)

	if err := h.authService.DisableTOTP(c.Context(), userID, req.Password); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "2FA disabled"})
}
