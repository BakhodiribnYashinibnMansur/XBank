package kafka

import (
	"context"
	"fmt"
	"time"

	notification "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/content/generic/notification/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/sse"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	kycpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/kyc"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// KYCEventHandler consumes KYC domain events from Kafka.
// Pushes real-time SSE notifications to connected clients.
type KYCEventHandler struct {
	sseHub *sse.Hub
}

// NewKYCEventHandler creates a handler that pushes SSE notifications
// for KYC events consumed from Kafka.
func NewKYCEventHandler(sseHub *sse.Hub) *KYCEventHandler {
	return &KYCEventHandler{sseHub: sseHub}
}

// Handle routes KYC events by topic to the appropriate processor.
func (h *KYCEventHandler) Handle(ctx context.Context, topic string, key []byte, value []byte) error {
	switch {
	case isKYCSubmitted(topic):
		return h.handleSubmitted(ctx, value)
	case isKYCApproved(topic):
		return h.handleApproved(ctx, value)
	case isKYCRejected(topic):
		return h.handleRejected(ctx, value)
	default:
		logger.Log.Warn("Unknown KYC topic", zap.String("topic", topic))
		return nil
	}
}

func (h *KYCEventHandler) handleSubmitted(ctx context.Context, data []byte) error {
	var msg kycpb.KYCSubmitted
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal KYCSubmitted: %w", err)
	}

	logger.Log.Info("Consumed KYCSubmitted",
		zap.String("verification_id", msg.VerificationId),
		zap.String("user_id", msg.UserId),
	)

	h.sseHub.Send(msg.UserId, notification.Event{
		ID:      uuid.NewString(),
		UserID:  msg.UserId,
		Type:    notification.EventKYCSubmitted,
		Title:   "KYC Submitted",
		Message: "Your KYC verification has been submitted and is under review.",
		Data: map[string]string{
			"verification_id": msg.VerificationId,
			"document_type":   msg.DocumentType,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

func (h *KYCEventHandler) handleApproved(ctx context.Context, data []byte) error {
	var msg kycpb.KYCApproved
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal KYCApproved: %w", err)
	}

	logger.Log.Info("Consumed KYCApproved",
		zap.String("verification_id", msg.VerificationId),
		zap.String("user_id", msg.UserId),
	)

	h.sseHub.Send(msg.UserId, notification.Event{
		ID:      uuid.NewString(),
		UserID:  msg.UserId,
		Type:    notification.EventKYCApproved,
		Title:   "KYC Approved",
		Message: "Your identity verification has been approved.",
		Data: map[string]string{
			"verification_id": msg.VerificationId,
			"level":           msg.Level,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

func (h *KYCEventHandler) handleRejected(ctx context.Context, data []byte) error {
	var msg kycpb.KYCRejected
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal KYCRejected: %w", err)
	}

	logger.Log.Warn("Consumed KYCRejected",
		zap.String("verification_id", msg.VerificationId),
		zap.String("user_id", msg.UserId),
		zap.String("reason", msg.Reason),
	)

	h.sseHub.Send(msg.UserId, notification.Event{
		ID:      uuid.NewString(),
		UserID:  msg.UserId,
		Type:    notification.EventKYCRejected,
		Title:   "KYC Rejected",
		Message: fmt.Sprintf("Your identity verification was rejected: %s", msg.Reason),
		Data: map[string]string{
			"verification_id": msg.VerificationId,
			"reason":          msg.Reason,
		},
		CreatedAt: time.Now(),
	})

	return nil
}

// Topic matchers
func isKYCSubmitted(t string) bool { return matchSuffix(t, "submitted") }
func isKYCApproved(t string) bool  { return matchSuffix(t, "approved") }
func isKYCRejected(t string) bool  { return matchSuffix(t, "rejected") }
