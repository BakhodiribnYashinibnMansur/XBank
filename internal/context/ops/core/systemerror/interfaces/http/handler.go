package http

import (
	"net/http"
	"strconv"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/systemerror/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/systemerror/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/core/systemerror/domain/repository"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	resolve *command.ResolveHandler
	list    *query.ListHandler
}

func NewHandler(r *command.ResolveHandler, l *query.ListHandler) *Handler {
	return &Handler{resolve: r, list: l}
}

func (h *Handler) Resolve(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if err := h.resolve.Handle(c.Context(), c.Params("id"), userID); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"resolved": true})
}

func (h *Handler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	result, err := h.list.Handle(c.Context(), repository.SystemErrorFilter{
		Severity: c.Query("severity"), Resolution: c.Query("resolution"), Code: c.Query("code"),
		Limit: limit, Offset: offset,
	})
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}
