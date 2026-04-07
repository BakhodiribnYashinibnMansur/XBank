package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/statistics/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	overview *query.OverviewHandler
}

func NewHandler(o *query.OverviewHandler) *Handler {
	return &Handler{overview: o}
}

func (h *Handler) Overview(c *fiber.Ctx) error {
	result, err := h.overview.Handle(c.Context())
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}
