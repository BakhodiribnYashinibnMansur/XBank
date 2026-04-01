package handler

import (
	"bufio"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/sse"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

type NotificationHandler struct {
	hub *sse.Hub
}

func NewNotificationHandler(hub *sse.Hub) *NotificationHandler {
	return &NotificationHandler{hub: hub}
}

// Stream godoc
// @Summary      SSE notification stream
// @Description  Server-Sent Events stream for real-time notifications
// @Tags         Notifications
// @Produce      text/event-stream
// @Security     BearerAuth
// @Router       /notifications/stream [get]
func (h *NotificationHandler) Stream(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return apperror.ErrUnauthorized
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ch := h.hub.Subscribe(userID)
		defer h.hub.Unsubscribe(userID, ch)

		// Send initial connection event
		fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
		w.Flush()

		for data := range ch {
			fmt.Fprintf(w, "event: notification\ndata: %s\n\n", data)
			if err := w.Flush(); err != nil {
				return // client disconnected
			}
		}
	})

	return nil
}

// compile check
var _ fasthttp.StreamWriter = func(w *bufio.Writer) {}
