package http

import (
	"net/http"
	"strconv"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/dataexport/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	request *command.RequestHandler
	process *command.ProcessHandler
	get     *query.GetHandler
	list    *query.ListHandler
}

func NewHandler(
	request *command.RequestHandler,
	process *command.ProcessHandler,
	get *query.GetHandler,
	list *query.ListHandler,
) *Handler {
	return &Handler{request: request, process: process, get: get, list: list}
}

func (h *Handler) Request(c *fiber.Ctx) error {
	var req application.RequestExportRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	id, err := h.request.Handle(c.Context(), req)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusCreated, fiber.Map{"id": id})
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

	result, err := h.list.Handle(c.Context(), domain.DataExportFilter{
		UserID: c.Query("user_id"),
		Status: c.Query("status"),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}
