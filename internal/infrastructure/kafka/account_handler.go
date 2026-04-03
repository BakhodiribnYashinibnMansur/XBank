package kafka

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/notification"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/sse"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	accountpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/accounts"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"time"
)

// AccountEventHandler consumes account domain events from Kafka.
// Responsibilities:
//   - Push real-time SSE notifications to connected clients
//   - Log consumed events for observability
//
// The read projection is updated synchronously in the application layer,
// so the consumer focuses on side-effects (notifications, analytics).
type AccountEventHandler struct {
	sseHub *sse.Hub
}

// NewAccountEventHandler creates a handler that pushes SSE notifications
// for account events consumed from Kafka.
func NewAccountEventHandler(sseHub *sse.Hub) *AccountEventHandler {
	return &AccountEventHandler{sseHub: sseHub}
}

// Handle routes account events by topic to the appropriate processor.
func (h *AccountEventHandler) Handle(ctx context.Context, topic string, key []byte, value []byte) error {
	switch {
	case isAccountOpened(topic):
		return h.handleOpened(ctx, value)
	case isAccountCredited(topic):
		return h.handleCredited(ctx, value)
	case isAccountDebited(topic):
		return h.handleDebited(ctx, value)
	case isAccountFrozen(topic):
		return h.handleFrozen(ctx, value)
	case isAccountClosed(topic):
		return h.handleClosed(ctx, value)
	default:
		logger.Log.Warn("Unknown account topic", zap.String("topic", topic))
		return nil
	}
}

func (h *AccountEventHandler) handleOpened(ctx context.Context, data []byte) error {
	var msg accountpb.AccountOpened
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal AccountOpened: %w", err)
	}

	logger.Log.Info("Consumed AccountOpened",
		zap.String("account_id", msg.AccountId),
		zap.String("user_id", msg.UserId),
	)

	h.sseHub.Send(msg.UserId, notification.Event{
		ID:        uuid.NewString(),
		UserID:    msg.UserId,
		Type:      "account.opened",
		Title:     "Account Created",
		Message:   fmt.Sprintf("Account %s created successfully", msg.AccountNumber),
		Data:      map[string]string{"account_id": msg.AccountId, "currency": msg.Currency},
		CreatedAt: time.Now(),
	})

	return nil
}

func (h *AccountEventHandler) handleCredited(ctx context.Context, data []byte) error {
	var msg accountpb.AccountCredited
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal AccountCredited: %w", err)
	}

	userID := msg.Metadata.GetUserId()
	logger.Log.Info("Consumed AccountCredited",
		zap.String("account_id", msg.AccountId),
		zap.Int64("amount", msg.Amount),
	)

	h.sseHub.Send(userID, notification.Event{
		ID:        uuid.NewString(),
		UserID:    userID,
		Type:      "account.credited",
		Title:     "Deposit Received",
		Message:   fmt.Sprintf("Deposited %d %s", msg.Amount, msg.Currency),
		Data:      map[string]string{"account_id": msg.AccountId, "amount": fmt.Sprintf("%d", msg.Amount)},
		CreatedAt: time.Now(),
	})

	return nil
}

func (h *AccountEventHandler) handleDebited(ctx context.Context, data []byte) error {
	var msg accountpb.AccountDebited
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal AccountDebited: %w", err)
	}

	userID := msg.Metadata.GetUserId()
	logger.Log.Info("Consumed AccountDebited",
		zap.String("account_id", msg.AccountId),
		zap.Int64("amount", msg.Amount),
	)

	h.sseHub.Send(userID, notification.Event{
		ID:        uuid.NewString(),
		UserID:    userID,
		Type:      "account.debited",
		Title:     "Withdrawal Processed",
		Message:   fmt.Sprintf("Withdrawn %d %s", msg.Amount, msg.Currency),
		Data:      map[string]string{"account_id": msg.AccountId, "amount": fmt.Sprintf("%d", msg.Amount)},
		CreatedAt: time.Now(),
	})

	return nil
}

func (h *AccountEventHandler) handleFrozen(ctx context.Context, data []byte) error {
	var msg accountpb.AccountFrozen
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal AccountFrozen: %w", err)
	}

	userID := msg.Metadata.GetUserId()
	logger.Log.Warn("Consumed AccountFrozen", zap.String("account_id", msg.AccountId))

	h.sseHub.Send(userID, notification.Event{
		ID:        uuid.NewString(),
		UserID:    userID,
		Type:      "account.frozen",
		Title:     "Account Frozen",
		Message:   "Your account has been frozen. Please contact support.",
		Data:      map[string]string{"account_id": msg.AccountId},
		CreatedAt: time.Now(),
	})

	return nil
}

func (h *AccountEventHandler) handleClosed(ctx context.Context, data []byte) error {
	var msg accountpb.AccountClosed
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("unmarshal AccountClosed: %w", err)
	}

	userID := msg.Metadata.GetUserId()
	logger.Log.Info("Consumed AccountClosed", zap.String("account_id", msg.AccountId))

	h.sseHub.Send(userID, notification.Event{
		ID:        uuid.NewString(),
		UserID:    userID,
		Type:      "account.closed",
		Title:     "Account Closed",
		Message:   "Your account has been closed.",
		Data:      map[string]string{"account_id": msg.AccountId},
		CreatedAt: time.Now(),
	})

	return nil
}

// Topic matchers — match by suffix so they work regardless of prefix config
func isAccountOpened(t string) bool    { return matchSuffix(t, "opened") }
func isAccountCredited(t string) bool  { return matchSuffix(t, "credited") }
func isAccountDebited(t string) bool   { return matchSuffix(t, "debited") }
func isAccountFrozen(t string) bool    { return matchSuffix(t, "frozen") }
func isAccountClosed(t string) bool    { return matchSuffix(t, "closed") }
