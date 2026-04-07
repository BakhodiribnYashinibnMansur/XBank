package http

import (
	"net/http"
	"strconv"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/core/notification/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	create   *command.CreateHandler
	markRead *command.MarkReadHandler
	delete   *command.DeleteHandler
	get      *query.GetHandler
	list     *query.ListHandler
}

func NewHandler(c *command.CreateHandler, mr *command.MarkReadHandler, d *command.DeleteHandler, g *query.GetHandler, l *query.ListHandler) *Handler {
	return &Handler{create: c, markRead: mr, delete: d, get: g, list: l}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	id, err := h.create.Handle(c.Context(), application.CreateNotificationRequest{
		UserID: req.UserID, Title: req.Title, Message: req.Message, Type: req.Type, Data: req.Data,
	})
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusCreated, fiber.Map{"id": id})
}

func (h *Handler) MarkAsRead(c *fiber.Ctx) error {
	if err := h.markRead.Handle(c.Context(), c.Params("id")); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"read": true})
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
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	userID, _ := c.Locals("user_id").(string)

	var unread *bool
	if u := c.Query("unread"); u == "true" {
		v := true
		unread = &v
	}

	result, err := h.list.Handle(c.Context(), domain.NotificationFilter{
		UserID: userID, Type: c.Query("type"), Unread: unread,
		Limit: limit, Offset: offset,
	})
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}
