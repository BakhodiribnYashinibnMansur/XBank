package http

import (
	"net/http"
	"strconv"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/file/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	upload *command.UploadHandler
	delete *command.DeleteHandler
	get    *query.GetHandler
	list   *query.ListHandler
}

func NewHandler(
	upload *command.UploadHandler,
	del *command.DeleteHandler,
	get *query.GetHandler,
	list *query.ListHandler,
) *Handler {
	return &Handler{upload: upload, delete: del, get: get, list: list}
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.get.Handle(c.Context(), id)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}

func (h *Handler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	result, err := h.list.Handle(c.Context(), domain.FileFilter{
		MimeType: c.Query("mime_type"),
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.delete.Handle(c.Context(), id); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"deleted": true})
}
