package transfer

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/account"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/transfer"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/config"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	commonpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/common"
	transferpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/transfers"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	transferRepo transfer.Repository
	eventRepo    transfer.EventRepository
	accountRepo  account.Repository
	txManager    shared.TxManager
	publisher    shared.EventPublisher
	topics       config.KafkaTopicsConfig
}

func NewService(
	transferRepo transfer.Repository,
	eventRepo transfer.EventRepository,
	accountRepo account.Repository,
	txManager shared.TxManager,
	publisher shared.EventPublisher,
	topics config.KafkaTopicsConfig,
) *Service {
	return &Service{
		transferRepo: transferRepo,
		eventRepo:    eventRepo,
		accountRepo:  accountRepo,
		txManager:    txManager,
		publisher:    publisher,
		topics:       topics,
	}
}

// Send - transfer funds between accounts (event sourced)
func (s *Service) Send(ctx context.Context, fromAccountID, toAccountID string, amount int64, currency shared.Currency, description string) (_ *transfer.Transfer, err error) {
	defer metrics.ObserveService("TransferService", "Send", time.Now(), &err)

	money, err := shared.NewMoney(amount, currency)
	if err != nil {
		return nil, err
	}

	tr, err := transfer.NewTransfer(fromAccountID, toAccountID, money, description)
	if err != nil {
		return nil, err
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Deadlock prevention: lock the smaller ID first
		firstID, secondID := fromAccountID, toAccountID
		if firstID > secondID {
			firstID, secondID = secondID, firstID
		}

		first, err := s.accountRepo.GetByIDForUpdate(txCtx, firstID)
		if err != nil {
			return account.ErrAccountNotFound
		}

		second, err := s.accountRepo.GetByIDForUpdate(txCtx, secondID)
		if err != nil {
			return account.ErrAccountNotFound
		}

		// Map back to from/to
		var fromAcc, toAcc *account.Account
		if firstID == fromAccountID {
			fromAcc, toAcc = first, second
		} else {
			fromAcc, toAcc = second, first
		}

		// Currency validation
		if fromAcc.Balance.Currency != currency || toAcc.Balance.Currency != currency {
			return shared.ErrCurrencyMismatch
		}

		// Domain operations (checks active status, sufficient funds, etc.)
		if err := fromAcc.Withdraw(money); err != nil {
			return err
		}
		if err := toAcc.Deposit(money); err != nil {
			return err
		}

		// Persist account changes
		if err := s.accountRepo.Update(txCtx, fromAcc); err != nil {
			return err
		}
		if err := s.accountRepo.Update(txCtx, toAcc); err != nil {
			return err
		}

		// Complete transfer and persist events + read projection
		tr.Complete()
		if err := s.eventRepo.Append(txCtx, tr.ID, 0, tr.UncommittedEvents()); err != nil {
			return err
		}
		if err := s.transferRepo.Create(txCtx, tr); err != nil {
			return err
		}
		tr.ClearUncommittedEvents()
		return nil
	})

	if err != nil {
		tr.Fail(err.Error())
		metrics.TransfersTotal.WithLabelValues("failed").Inc()
		s.publishTransferFailed(ctx, tr, err.Error())
		return tr, err
	}

	metrics.TransfersTotal.WithLabelValues("completed").Inc()
	s.publishTransferCompleted(ctx, tr)
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

// --- Kafka publish helpers (best-effort, after DB commit) ---

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
