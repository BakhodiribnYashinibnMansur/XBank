package http

import (
	"net/http"
	"strconv"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/supporting/announcement/domain/repository"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	create     *command.CreateHandler
	publish    *command.PublishHandler
	delete     *command.DeleteHandler
	get        *query.GetHandler
	list       *query.ListHandler
	listActive *query.ListActiveHandler
}

func NewHandler(c *command.CreateHandler, p *command.PublishHandler, d *command.DeleteHandler, g *query.GetHandler, l *query.ListHandler, la *query.ListActiveHandler) *Handler {
	return &Handler{create: c, publish: p, delete: d, get: g, list: l, listActive: la}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req application.CreateAnnouncementRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	id, err := h.create.Handle(c.Context(), req)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusCreated, fiber.Map{"id": id})
}

func (h *Handler) Publish(c *fiber.Ctx) error {
	if err := h.publish.Handle(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"published": true})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	if err := h.delete.Handle(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"deleted": true})
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	r, err := h.get.Handle(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, r)
}

func (h *Handler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	r, err := h.list.Handle(c.Context(), repository.AnnouncementFilter{
		Status: c.Query("status"), Limit: limit, Offset: offset,
	})
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, r)
}

func (h *Handler) ListActive(c *fiber.Ctx) error {
	r, err := h.listActive.Handle(c.Context())
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, r)
}
