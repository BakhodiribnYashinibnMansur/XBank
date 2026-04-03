package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/notification"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/sse"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	transferpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/transfers"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// TransferEventHandler consumes transfer domain events from Kafka
// and pushes SSE notifications to both sender and receiver.
type TransferEventHandler struct {
	sseHub *sse.Hub
}

// NewTransferEventHandler creates a handler for transfer events.
func NewTransferEventHandler(sseHub *sse.Hub) *TransferEventHandler {
	return &TransferEventHandler{sseHub: sseHub}
}

// Handle dispatches transfer events by topic.
func (h *TransferEventHandler) Handle(ctx context.Context, topic string, key []byte, value []byte) error {
	switch {
	case isTransferCompleted(topic):
		return h.handleCompleted(ctx, value)
	case isTransferFailed(topic):
		return h.handleFailed(ctx, value)
	case isTransferCreated(topic):
		return h.handleCreated(ctx, value)
	default:
		logger.Log.Warn("Unknown transfer topic", zap.String("topic", topic))
		return nil
	}
}

func (h *TransferEventHandler) handleCreated(ctx context.Context, data []byte) error {
	var msg transferpb.TransferCreated
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal TransferCreated: %w", err)
	}

	userID := msg.Metadata.GetUserId()
	logger.Log.Info("Consumed TransferCreated",
		zap.String("transfer_id", msg.TransferId),
		zap.Int64("amount", msg.Amount),
	)

	h.sseHub.Send(userID, notification.Event{
		ID:     uuid.NewString(),
		UserID: userID,
		Type:   "transfer.created",
		Title:  "Transfer Initiated",
		Message: fmt.Sprintf("Transfer of %d %s initiated", msg.Amount, msg.Currency),
		Data: map[string]string{
			"transfer_id": msg.TransferId,
			"amount":      fmt.Sprintf("%d", msg.Amount),
			"currency":    msg.Currency,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

func (h *TransferEventHandler) handleCompleted(ctx context.Context, data []byte) error {
	var msg transferpb.TransferCompleted
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal TransferCompleted: %w", err)
	}

	userID := msg.Metadata.GetUserId()
	logger.Log.Info("Consumed TransferCompleted",
		zap.String("transfer_id", msg.TransferId),
		zap.Int64("amount", msg.Amount),
	)

	// Notify sender
	h.sseHub.Send(userID, notification.Event{
		ID:      uuid.NewString(),
		UserID:  userID,
		Type:    notification.EventTransferCompleted,
		Title:   "Transfer Completed",
		Message: fmt.Sprintf("Transfer of %d %s completed successfully", msg.Amount, msg.Currency),
		Data: map[string]string{
			"transfer_id": msg.TransferId,
			"amount":      fmt.Sprintf("%d", msg.Amount),
			"currency":    msg.Currency,
		},
		CreatedAt: time.Now(),
	})

	// Notify receiver (different account owner)
	h.sseHub.Send(msg.ToAccountId, notification.Event{
		ID:      uuid.NewString(),
		UserID:  msg.ToAccountId,
		Type:    notification.EventTransferReceived,
		Title:   "Funds Received",
		Message: fmt.Sprintf("Received %d %s", msg.Amount, msg.Currency),
		Data: map[string]string{
			"transfer_id": msg.TransferId,
			"amount":      fmt.Sprintf("%d", msg.Amount),
			"currency":    msg.Currency,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

func (h *TransferEventHandler) handleFailed(ctx context.Context, data []byte) error {
	var msg transferpb.TransferFailed
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal TransferFailed: %w", err)
	}

	userID := msg.Metadata.GetUserId()
	logger.Log.Warn("Consumed TransferFailed",
		zap.String("transfer_id", msg.TransferId),
		zap.String("reason", msg.Reason),
	)

	h.sseHub.Send(userID, notification.Event{
		ID:      uuid.NewString(),
		UserID:  userID,
		Type:    notification.EventTransferFailed,
		Title:   "Transfer Failed",
		Message: fmt.Sprintf("Transfer of %d %s failed: %s", msg.Amount, msg.Currency, msg.Reason),
		Data: map[string]string{
			"transfer_id": msg.TransferId,
			"amount":      fmt.Sprintf("%d", msg.Amount),
			"reason":      msg.Reason,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

func isTransferCreated(t string) bool   { return matchSuffix(t, "created") }
func isTransferCompleted(t string) bool { return matchSuffix(t, "completed") }
func isTransferFailed(t string) bool    { return matchSuffix(t, "failed") }
