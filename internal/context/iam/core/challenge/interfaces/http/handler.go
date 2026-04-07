package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/core/challenge/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/core/challenge/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *command.Service
}

func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Request(c *fiber.Ctx) error {
	var req ChallengeRequestDTO
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	if req.Action == "" {
		return apperror.ErrMissingField.WithMessage("action is required")
	}

	method := domain.MethodPassword
	if req.Method == string(domain.MethodTOTP) {
		method = domain.MethodTOTP
	}

	userID := c.Locals("user_id").(string)

	ch, err := h.service.Request(c.Context(), userID, method, req.Action, req.Metadata)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, ChallengeResponse{
		ID:        ch.ID,
		Method:    string(ch.Method),
		Status:    string(ch.Status),
		Action:    ch.Action,
		ExpiresAt: ch.ExpiresAt,
	})
}

func (h *Handler) Verify(c *fiber.Ctx) error {
	var req ChallengeVerifyDTO
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

	return apperror.Success(c, http.StatusOK, ChallengeVerifiedResponse{
		ChallengeID: ch.ID,
		Token:       ch.Token,
		ExpiresAt:   ch.ExpiresAt,
	})
}
