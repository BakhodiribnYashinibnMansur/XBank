package command

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/contract/ports"
	transfer "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/outbox"
	commonpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/common"
	transferpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/transfers"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	transferRepo transfer.Repository
	eventRepo    transfer.EventRepository
	accountPort  ports.AccountTransferPort
	txManager    domain.TxManager
	publisher    domain.EventPublisher
	topics       config.KafkaTopicsConfig
}

func NewService(
	transferRepo transfer.Repository,
	eventRepo transfer.EventRepository,
	accountPort ports.AccountTransferPort,
	txManager domain.TxManager,
	publisher domain.EventPublisher,
	topics config.KafkaTopicsConfig,
) *Service {
	return &Service{
		transferRepo: transferRepo,
		eventRepo:    eventRepo,
		accountPort:  accountPort,
		txManager:    txManager,
		publisher:    publisher,
		topics:       topics,
	}
}

// Send - transfer funds between accounts (event sourced)
func (s *Service) Send(ctx context.Context, fromAccountID, toAccountID string, amount int64, currency domain.Currency, description string) (_ *transfer.Transfer, err error) {
	defer metrics.ObserveService("TransferService", "Send", time.Now(), &err)

	money, err := domain.NewMoney(amount, currency)
	if err != nil {
		return nil, err
	}

	tr, err := transfer.NewTransfer(fromAccountID, toAccountID, money, description)
	if err != nil {
		return nil, err
	}

	err = s.txManager.WithSerializableTx(ctx, func(txCtx context.Context) error {
		// Delegate account debit/credit to the Account BC via port
		if err := s.accountPort.TransferFunds(txCtx, fromAccountID, toAccountID, amount, string(currency)); err != nil {
			return err
		}

		// Complete transfer and persist events + read projection
		tr.Complete()
		if err := s.eventRepo.Append(txCtx, tr.ID, tr.UncommittedEvents()); err != nil {
			return err
		}
		if err := s.transferRepo.Create(txCtx, tr); err != nil {
			return err
		}
		tr.ClearUncommittedEvents()

		// Outbox: publish within the same transaction
		txCtx = outbox.WithOutboxMeta(txCtx, "Transfer", tr.ID)
		s.publishTransferCompleted(txCtx, tr)
		return nil
	})

	if err != nil {
		tr.Fail(err.Error())
		metrics.TransfersTotal.WithLabelValues("failed").Inc()
		return tr, err
	}

	metrics.TransfersTotal.WithLabelValues("completed").Inc()
	return tr, nil
}

// GetByID - get a transfer by ID
func (s *Service) GetByID(ctx context.Context, id string) (_ *transfer.Transfer, err error) {
	defer metrics.ObserveService("TransferService", "GetByID", time.Now(), &err)
	return s.transferRepo.GetByID(ctx, id)
}

// ListByAccountID - get transfers for an account with pagination
func (s *Service) ListByAccountID(ctx context.Context, accountID string, limit, offset int) (_ []*transfer.Transfer, _ int64, err error) {
	defer metrics.ObserveService("TransferService", "ListByAccountID", time.Now(), &err)

	transfers, err := s.transferRepo.ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.transferRepo.CountByAccountID(ctx, accountID)
	if err != nil {
		return nil, 0, err
	}
	return transfers, total, nil
}

// GetHistory - load all domain events for a transfer
func (s *Service) GetHistory(ctx context.Context, transferID string) (_ []transfer.Event, err error) {
	defer metrics.ObserveService("TransferService", "GetHistory", time.Now(), &err)

	events, err := s.eventRepo.LoadEvents(ctx, transferID)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, transfer.ErrTransferNotFound
	}
	return events, nil
}

// --- Kafka publish helpers (outbox: called within DB transaction) ---

func newMetadata() *commonpb.Metadata {
	return &commonpb.Metadata{
		EventId:   uuid.New().String(),
		Timestamp: timestamppb.Now(),
		Source:    "xbank-api",
	}
}

func (s *Service) publishTransferCompleted(ctx context.Context, tr *transfer.Transfer) {
	msg := &transferpb.TransferCompleted{
		Metadata:      newMetadata(),
		TransferId:    tr.ID,
		FromAccountId: tr.FromAccountID,
		ToAccountId:   tr.ToAccountID,
		Amount:        tr.Amount.Amount,
		Currency:      string(tr.Amount.Currency),
	}
	if data, err := proto.Marshal(msg); err == nil {
		s.publisher.Publish(ctx, s.topics.TransferCompleted, tr.FromAccountID, data)
	}
}

func (s *Service) publishTransferFailed(ctx context.Context, tr *transfer.Transfer, reason string) {
	msg := &transferpb.TransferFailed{
		Metadata:      newMetadata(),
		TransferId:    tr.ID,
		FromAccountId: tr.FromAccountID,
		ToAccountId:   tr.ToAccountID,
		Amount:        tr.Amount.Amount,
		Currency:      string(tr.Amount.Currency),
		Reason:        reason,
	}
	if data, err := proto.Marshal(msg); err == nil {
		s.publisher.Publish(ctx, s.topics.TransferFailed, tr.FromAccountID, data)
	}
}
