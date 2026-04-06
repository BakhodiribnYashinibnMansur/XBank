package http

import (
	"net/http"
	"strconv"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/translation/domain/repository"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	create *command.CreateHandler
	update *command.UpdateHandler
	delete *command.DeleteHandler
	get    *query.GetHandler
	list   *query.ListHandler
}

func NewHandler(c *command.CreateHandler, u *command.UpdateHandler, d *command.DeleteHandler, g *query.GetHandler, l *query.ListHandler) *Handler {
	return &Handler{create: c, update: u, delete: d, get: g, list: l}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateTranslationRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	id, err := h.create.Handle(c.Context(), application.CreateTranslationRequest{
		Key: req.Key, Language: req.Language, Value: req.Value, Group: req.Group,
	})
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusCreated, fiber.Map{"id": id})
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateTranslationRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if err := h.update.Handle(c.Context(), id, req.Value); err != nil {
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

func (h *Handler) GetByID(c *fiber.Ctx) error {
	result, err := h.get.Handle(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}

func (h *Handler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	result, err := h.list.Handle(c.Context(), repository.TranslationFilter{
		Language: c.Query("lang"), Group: c.Query("group"), Key: c.Query("key"),
		Limit: limit, Offset: offset,
	})
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}
