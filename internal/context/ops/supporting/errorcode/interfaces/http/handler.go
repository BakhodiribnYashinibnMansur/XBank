package http

import (
	"net/http"
	"strconv"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/errorcode/domain/repository"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	create *command.CreateHandler
	update *command.UpdateHandler
	delete *command.DeleteHandler
	list   *query.ListHandler
	lookup *query.LookupHandler
}

func NewHandler(c *command.CreateHandler, u *command.UpdateHandler, d *command.DeleteHandler, l *query.ListHandler, lu *query.LookupHandler) *Handler {
	return &Handler{create: c, update: u, delete: d, list: l, lookup: lu}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req application.CreateErrorCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	id, err := h.create.Handle(c.Context(), req)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusCreated, fiber.Map{"id": id})
}

func (h *Handler) Update(c *fiber.Ctx) error {
	var req application.UpdateErrorCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if err := h.update.Handle(c.Context(), c.Params("id"), req); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"updated": true})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.delete.Handle(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"deleted": true})
}

func (h *Handler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	result, err := h.list.Handle(c.Context(), repository.ErrorCodeFilter{
		Code: c.Query("code"), Category: c.Query("category"), Severity: c.Query("severity"),
		Limit: limit, Offset: offset,
	})
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}

func (h *Handler) Lookup(c *fiber.Ctx) error {
	result, err := h.lookup.Handle(c.Context(), c.Params("code"))
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}
