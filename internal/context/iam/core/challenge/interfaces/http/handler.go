package http

import (
	"net/http"

	challengeApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/challenge"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/challenge"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type ChallengeHandler struct {
	service *challengeApp.Service
}

func NewChallengeHandler(service *challengeApp.Service) *ChallengeHandler {
	return &ChallengeHandler{service: service}
}

// Request godoc
// @Summary      Request a step-up authentication challenge
// @Description  Creates a new challenge that must be verified before performing sensitive operations
// @Tags         Challenge
// @Accept       json
// @Produce      json
// @Param        body body dto.ChallengeRequestDTO true "Challenge request"
// @Success      201 {object} dto.ChallengeResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      429 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /auth/challenge/request [post]
func (h *ChallengeHandler) Request(c *fiber.Ctx) error {
	var req dto.ChallengeRequestDTO
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	if req.Action == "" {
		return apperror.ErrMissingField.WithMessage("action is required")
	}

	method := challenge.MethodPassword
	if req.Method == string(challenge.MethodTOTP) {
		method = challenge.MethodTOTP
	}

	userID := c.Locals("user_id").(string)

	ch, err := h.service.Request(c.Context(), userID, method, req.Action, req.Metadata)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, dto.ChallengeResponse{
		ID:        ch.ID,
		Method:    string(ch.Method),
		Status:    string(ch.Status),
		Action:    ch.Action,
		ExpiresAt: ch.ExpiresAt,
	})
}

// Verify godoc
// @Summary      Verify a step-up authentication challenge
// @Description  Verifies identity (password) and returns a challenge token for sensitive operations
// @Tags         Challenge
// @Accept       json
// @Produce      json
// @Param        body body dto.ChallengeVerifyDTO true "Verification details"
// @Success      200 {object} dto.ChallengeVerifiedResponse
// @Failure      401 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /auth/challenge/verify [post]
func (h *ChallengeHandler) Verify(c *fiber.Ctx) error {
	var req dto.ChallengeVerifyDTO
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	if req.ChallengeID == "" || req.Password == "" {
		return apperror.ErrMissingField.WithMessage("challenge_id and password are required")
	}

	userID := c.Locals("user_id").(string)

	ch, err := h.service.Verify(c.Context(), req.ChallengeID, userID, req.Password)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, dto.ChallengeVerifiedResponse{
		ChallengeID: ch.ID,
		Token:       ch.Token,
		ExpiresAt:   ch.ExpiresAt,
	})
}
