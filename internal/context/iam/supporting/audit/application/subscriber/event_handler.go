package subscriber

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/audit/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	accountpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/accounts"
	transferpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/transfers"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// AuditEventHandler consumes Kafka events and writes audit log entries.
type AuditEventHandler struct {
	service *command.Service
}

func NewAuditEventHandler(service *command.Service) *AuditEventHandler {
	return &AuditEventHandler{service: service}
}

// Handle routes events by topic to create audit log entries.
func (h *AuditEventHandler) Handle(ctx context.Context, topic string, key []byte, value []byte) error {
	switch {
	case matchSuffix(topic, "opened"):
		return h.handleAccountOpened(ctx, value)
	case matchSuffix(topic, "credited"):
		return h.handleAccountCredited(ctx, value)
	case matchSuffix(topic, "debited"):
		return h.handleAccountDebited(ctx, value)
	case matchSuffix(topic, "transfer.created"):
		return h.handleTransferCreated(ctx, value)
	case matchSuffix(topic, "transfer.completed"):
		return h.handleTransferCompleted(ctx, value)
	case matchSuffix(topic, "transfer.failed"):
		return h.handleTransferFailed(ctx, value)
	default:
		logger.Log.Debug("Audit: unhandled topic", zap.String("topic", topic))
		return nil
	}
}

func (h *AuditEventHandler) handleAccountOpened(ctx context.Context, data []byte) error {
	var msg accountpb.AccountOpened
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("audit: unmarshal AccountOpened: %w", err)
	}
	_, err := h.service.CreateAuditLog(ctx, command.CreateAuditLogInput{
		AggregateType: "Account",
		AggregateID:   msg.AccountId,
		Action:        "AccountOpened",
		ActorID:       msg.UserId,
		Attributes:    map[string]any{"account_number": msg.AccountNumber, "currency": msg.Currency},
	})
	return err
}

func (h *AuditEventHandler) handleAccountCredited(ctx context.Context, data []byte) error {
	var msg accountpb.AccountCredited
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("audit: unmarshal AccountCredited: %w", err)
	}
	_, err := h.service.CreateAuditLog(ctx, command.CreateAuditLogInput{
		AggregateType: "Account",
		AggregateID:   msg.AccountId,
		Action:        "AccountCredited",
		ActorID:       msg.Metadata.GetUserId(),
		Attributes:    map[string]any{"amount": msg.Amount, "currency": msg.Currency},
	})
	return err
}

func (h *AuditEventHandler) handleAccountDebited(ctx context.Context, data []byte) error {
	var msg accountpb.AccountDebited
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("audit: unmarshal AccountDebited: %w", err)
	}
	_, err := h.service.CreateAuditLog(ctx, command.CreateAuditLogInput{
		AggregateType: "Account",
		AggregateID:   msg.AccountId,
		Action:        "AccountDebited",
		ActorID:       msg.Metadata.GetUserId(),
		Attributes:    map[string]any{"amount": msg.Amount, "currency": msg.Currency},
	})
	return err
}

func (h *AuditEventHandler) handleTransferCreated(ctx context.Context, data []byte) error {
	var msg transferpb.TransferCreated
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("audit: unmarshal TransferCreated: %w", err)
	}
	_, err := h.service.CreateAuditLog(ctx, command.CreateAuditLogInput{
		AggregateType: "Transfer",
		AggregateID:   msg.TransferId,
		Action:        "TransferCreated",
		ActorID:       msg.Metadata.GetUserId(),
		Attributes:    map[string]any{"from": msg.FromAccountId, "to": msg.ToAccountId, "amount": msg.Amount},
	})
	return err
}

func (h *AuditEventHandler) handleTransferCompleted(ctx context.Context, data []byte) error {
	var msg transferpb.TransferCompleted
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("audit: unmarshal TransferCompleted: %w", err)
	}
	_, err := h.service.CreateAuditLog(ctx, command.CreateAuditLogInput{
		AggregateType: "Transfer",
		AggregateID:   msg.TransferId,
		Action:        "TransferCompleted",
		ActorID:       msg.Metadata.GetUserId(),
	})
	return err
}

func (h *AuditEventHandler) handleTransferFailed(ctx context.Context, data []byte) error {
	var msg transferpb.TransferFailed
	if err := proto.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf("audit: unmarshal TransferFailed: %w", err)
	}
	_, err := h.service.CreateAuditLog(ctx, command.CreateAuditLogInput{
		AggregateType: "Transfer",
		AggregateID:   msg.TransferId,
		Action:        "TransferFailed",
		ActorID:       msg.Metadata.GetUserId(),
		Attributes:    map[string]any{"reason": msg.Reason},
	})
	return err
}

func matchSuffix(topic, suffix string) bool {
	if len(topic) < len(suffix) {
		return false
	}
	return topic[len(topic)-len(suffix):] == suffix
}
