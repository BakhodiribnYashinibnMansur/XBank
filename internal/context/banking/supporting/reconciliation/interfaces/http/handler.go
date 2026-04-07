package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/reconciliation/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *command.Service
}

func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

// CheckAccount godoc
// @Summary      Reconcile a single account (admin)
// @Tags         Admin - Reconciliation
// @Produce      json
// @Param        account_id query string true "Account ID"
// @Success      200 {object} command.Result
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/reconciliation/check [get]
func (h *Handler) CheckAccount(c *fiber.Ctx) error {
	accountID := c.Query("account_id")
	if accountID == "" {
		return apperror.ErrMissingField.WithMessage("account_id is required")
	}

	result, err := h.service.CheckAccount(c.Context(), accountID)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}

// CheckAllByUser godoc
// @Summary      Reconcile all accounts for a user (admin)
// @Tags         Admin - Reconciliation
// @Produce      json
// @Param        user_id query string true "User ID"
// @Success      200 {object} map[string]any
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/reconciliation/check-all [get]
func (h *Handler) CheckAllByUser(c *fiber.Ctx) error {
	userID := c.Query("user_id")
	if userID == "" {
		return apperror.ErrMissingField.WithMessage("user_id is required")
	}

	results, err := h.service.CheckAllAccounts(c.Context(), userID)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{
		"results": results,
		"summary": command.Summary(results),
	})
}
