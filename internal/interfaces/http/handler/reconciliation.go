package handler

import (
	"net/http"

	reconApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/reconciliation"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type ReconciliationHandler struct {
	service *reconApp.Service
}

func NewReconciliationHandler(service *reconApp.Service) *ReconciliationHandler {
	return &ReconciliationHandler{service: service}
}

// CheckAccount godoc
// @Summary      Reconcile a single account (admin)
// @Tags         Admin - Reconciliation
// @Param        account_id query string true "Account ID"
// @Success      200 {object} reconciliation.Result
// @Security     BearerAuth
// @Router       /admin/reconciliation/check [get]
func (h *ReconciliationHandler) CheckAccount(c *fiber.Ctx) error {
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
// @Param        user_id query string true "User ID"
// @Success      200 {object} map[string]interface{}
// @Security     BearerAuth
// @Router       /admin/reconciliation/check-all [get]
func (h *ReconciliationHandler) CheckAllByUser(c *fiber.Ctx) error {
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
		"summary": reconApp.Summary(results),
	})
}
