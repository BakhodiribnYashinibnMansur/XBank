package http

import (
	"net/http"
	"strconv"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/usersetting/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	upsert *command.UpsertHandler
	delete *command.DeleteHandler
	list   *query.ListHandler
}

func NewHandler(
	upsert *command.UpsertHandler,
	del *command.DeleteHandler,
	list *query.ListHandler,
) *Handler {
	return &Handler{upsert: upsert, delete: del, list: list}
}

func (h *Handler) Upsert(c *fiber.Ctx) error {
	var req application.UpsertSettingRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	// Use authenticated user's ID
	if userID, ok := c.Locals("user_id").(string); ok && req.UserID == "" {
		req.UserID = userID
	}

	if err := h.upsert.Handle(c.Context(), req); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"updated": true})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.delete.Handle(c.Context(), id); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"deleted": true})
}

func (h *Handler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	userID := c.Query("user_id")
	if userID == "" {
		if uid, ok := c.Locals("user_id").(string); ok {
			userID = uid
		}
	}

	result, err := h.list.Handle(c.Context(), domain.UserSettingFilter{
		UserID: userID,
		Key:    c.Query("key"),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}
