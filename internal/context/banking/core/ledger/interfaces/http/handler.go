package http

import (
	"net/http"
	"strconv"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	list *query.ListHandler
}

func NewHandler(list *query.ListHandler) *Handler {
	return &Handler{list: list}
}

func (h *Handler) ListByAccount(c *fiber.Ctx) error {
	accountID := c.Query("account_id")
	if accountID == "" {
		return apperror.ErrMissingField.WithMessage("account_id is required")
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	result, err := h.list.Handle(c.Context(), accountID, limit, offset)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}
