package kafka

import (
	"context"
	"fmt"
	"time"

	notification "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/notification/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/sse"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	notifpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/notifications"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// NotificationEventHandler consumes notification domain events from Kafka.
// Pushes real-time SSE notifications to connected clients and can trigger
// further delivery channels (push, email, SMS).
type NotificationEventHandler struct {
	sseHub *sse.Hub
}

// NewNotificationEventHandler creates a handler for notification events.
func NewNotificationEventHandler(sseHub *sse.Hub) *NotificationEventHandler {
	return &NotificationEventHandler{sseHub: sseHub}
}

// Handle dispatches notification events by topic.
func (h *NotificationEventHandler) Handle(ctx context.Context, topic string, key []byte, value []byte) error {
	switch {
	case isNotificationRequested(topic):
		return h.handleRequested(ctx, value)
	default:
		logger.Log.Warn("Unknown notification topic", zap.String("topic", topic))
		return nil
	}
}

func (h *NotificationEventHandler) handleRequested(ctx context.Context, data []byte) error {
	var msg notifpb.NotificationRequested
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal NotificationRequested: %w", err)
	}

	logger.Log.Info("Consumed NotificationRequested",
		zap.String("notification_id", msg.NotificationId),
		zap.String("user_id", msg.UserId),
		zap.String("type", msg.Type),
	)

	h.sseHub.Send(msg.UserId, notification.Event{
		ID:      uuid.NewString(),
		UserID:  msg.UserId,
		Type:    notification.EventType(msg.Type),
		Title:   msg.Title,
		Message: msg.Body,
		Data:    msg.Data,
		CreatedAt: time.Now(),
	})

	return nil
}

// Topic matcher
func isNotificationRequested(t string) bool { return matchSuffix(t, "requested") }
