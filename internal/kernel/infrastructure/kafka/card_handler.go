package kafka

import (
	"context"
	"fmt"
	"time"

	notification "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/notification/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/sse"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	cardpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/cards"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// CardEventHandler consumes card domain events from Kafka.
// Pushes real-time SSE notifications to connected clients.
type CardEventHandler struct {
	sseHub *sse.Hub
}

// NewCardEventHandler creates a handler that pushes SSE notifications
// for card events consumed from Kafka.
func NewCardEventHandler(sseHub *sse.Hub) *CardEventHandler {
	return &CardEventHandler{sseHub: sseHub}
}

// Handle routes card events by topic to the appropriate processor.
func (h *CardEventHandler) Handle(ctx context.Context, topic string, key []byte, value []byte) error {
	switch {
	case isCardIssued(topic):
		return h.handleIssued(ctx, value)
	case isCardBlocked(topic):
		return h.handleBlocked(ctx, value)
	case isCardActivated(topic):
		return h.handleActivated(ctx, value)
	default:
		logger.Log.Warn("Unknown card topic", zap.String("topic", topic))
		return nil
	}
}

func (h *CardEventHandler) handleIssued(ctx context.Context, data []byte) error {
	var msg cardpb.CardIssued
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal CardIssued: %w", err)
	}

	userID := msg.Metadata.GetUserId()
	logger.Log.Info("Consumed CardIssued",
		zap.String("card_id", msg.CardId),
		zap.String("account_id", msg.AccountId),
	)

	h.sseHub.Send(userID, notification.Event{
		ID:     uuid.NewString(),
		UserID: userID,
		Type:   notification.EventCardIssued,
		Title:  "New Card Issued",
		Message: fmt.Sprintf("A new %s card has been issued (ending %s)",
			msg.CardType, lastFour(msg.MaskedPan)),
		Data: map[string]string{
			"card_id":    msg.CardId,
			"account_id": msg.AccountId,
			"card_type":  msg.CardType,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

func (h *CardEventHandler) handleBlocked(ctx context.Context, data []byte) error {
	var msg cardpb.CardBlocked
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal CardBlocked: %w", err)
	}

	userID := msg.Metadata.GetUserId()
	logger.Log.Warn("Consumed CardBlocked",
		zap.String("card_id", msg.CardId),
		zap.String("reason", msg.Reason),
	)

	h.sseHub.Send(userID, notification.Event{
		ID:      uuid.NewString(),
		UserID:  userID,
		Type:    notification.EventCardBlocked,
		Title:   "Card Blocked",
		Message: fmt.Sprintf("Your card has been blocked: %s", msg.Reason),
		Data: map[string]string{
			"card_id": msg.CardId,
			"reason":  msg.Reason,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

func (h *CardEventHandler) handleActivated(ctx context.Context, data []byte) error {
	var msg cardpb.CardActivated
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal CardActivated: %w", err)
	}

	userID := msg.Metadata.GetUserId()
	logger.Log.Info("Consumed CardActivated", zap.String("card_id", msg.CardId))

	h.sseHub.Send(userID, notification.Event{
		ID:      uuid.NewString(),
		UserID:  userID,
		Type:    notification.EventCardActivated,
		Title:   "Card Activated",
		Message: "Your card has been activated and is ready to use.",
		Data: map[string]string{
			"card_id": msg.CardId,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

// Topic matchers
func isCardIssued(t string) bool    { return matchSuffix(t, "issued") }
func isCardBlocked(t string) bool   { return matchSuffix(t, "blocked") }
func isCardActivated(t string) bool { return matchSuffix(t, "activated") }

// lastFour returns the last 4 characters of a masked PAN for display.
func lastFour(pan string) string {
	if len(pan) < 4 {
		return pan
	}
	return pan[len(pan)-4:]
}
