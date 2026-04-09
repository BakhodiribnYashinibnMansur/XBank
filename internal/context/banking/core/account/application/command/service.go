package command

import (
	"context"
	"fmt"
	"time"

	account "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/outbox"
	accountpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/accounts"
	commonpb "github.com/BakhodiribnYashinibnMansur/XBank/proto/common"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const snapshotInterval = 100

type Service struct {
	repo      account.Repository
	eventRepo account.EventRepository
	txManager domain.TxManager
	publisher domain.EventPublisher
	topics    config.KafkaTopicsConfig
	auditLog  domain.AuditLog
}

func NewService(
	repo account.Repository,
	eventRepo account.EventRepository,
	txManager domain.TxManager,
	publisher domain.EventPublisher,
	topics config.KafkaTopicsConfig,
	auditLog domain.AuditLog,
) *Service {
	return &Service{
		repo:      repo,
		eventRepo: eventRepo,
		txManager: txManager,
		publisher: publisher,
		topics:    topics,
		auditLog:  auditLog,
	}
}

// CreateAccount - open a new account (event sourced)
func (s *Service) CreateAccount(ctx context.Context, userID string, currency domain.Currency) (result *account.Account, err error) {
	defer metrics.ObserveService("AccountService", "CreateAccount", time.Now(), &err)

	acc, err := account.NewAccount(userID, currency)
	if err != nil {
		return nil, err
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.eventRepo.Append(txCtx, acc.ID, acc.UncommittedEvents()); err != nil {
			return err
		}
		if err := s.repo.Create(txCtx, acc); err != nil {
			return err
		}
		acc.ClearUncommittedEvents()

		// Outbox: publish within the same transaction
		txCtx = outbox.WithOutboxMeta(txCtx, "Account", acc.ID)
		s.publishAccountOpened(txCtx, acc, userID)
		return nil
	})
	if err != nil {
		return nil, err
	}

	metrics.AccountsCreatedTotal.Inc()
	s.audit(ctx, acc.ID, "AccountOpened", map[string]string{
		"user_id": userID, "currency": string(currency), "account_number": acc.AccountNumber,
	})
	return acc, nil
}

// GetByID - read from projection (fast)
func (s *Service) GetByID(ctx context.Context, id string) (result *account.Account, err error) {
	defer metrics.ObserveService("AccountService", "GetByID", time.Now(), &err)
	return s.repo.GetByID(ctx, id)
}

// ListByUserID - read from projection (paginated)
func (s *Service) ListByUserID(ctx context.Context, userID string, limit, offset int) (_ []*account.Account, _ int64, err error) {
	defer metrics.ObserveService("AccountService", "ListByUserID", time.Now(), &err)

	accounts, err := s.repo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	return accounts, total, nil
}

// Deposit - event sourced deposit
func (s *Service) Deposit(ctx context.Context, accountID string, amount int64) (_ *account.Account, err error) {
	defer metrics.ObserveService("AccountService", "Deposit", time.Now(), &err)

	var result *account.Account

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		acc, err := s.loadAggregate(txCtx, accountID)
		if err != nil {
			return err
		}

		money, err := domain.NewMoney(amount, acc.Balance.Currency)
		if err != nil {
			return err
		}
		if err := acc.Deposit(money); err != nil {
			return err
		}

		if err := s.saveAggregate(txCtx, acc); err != nil {
			return err
		}

		// Outbox: publish within the same transaction
		txCtx = outbox.WithOutboxMeta(txCtx, "Account", acc.ID)
		s.publishAccountCredited(txCtx, acc, amount)

		result = acc
		return nil
	})
	if err != nil {
		return nil, err
	}

	metrics.DepositsTotal.Inc()
	s.audit(ctx, result.ID, "Credited", map[string]string{
		"amount": fmt.Sprintf("%d", amount), "currency": string(result.Balance.Currency),
	})
	return result, nil
}

// Withdraw - event sourced withdrawal
func (s *Service) Withdraw(ctx context.Context, accountID string, amount int64) (_ *account.Account, err error) {
	defer metrics.ObserveService("AccountService", "Withdraw", time.Now(), &err)

	var result *account.Account

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		acc, err := s.loadAggregate(txCtx, accountID)
		if err != nil {
			return err
		}

		money, err := domain.NewMoney(amount, acc.Balance.Currency)
		if err != nil {
			return err
		}
		if err := acc.Withdraw(money); err != nil {
			return err
		}

		if err := s.saveAggregate(txCtx, acc); err != nil {
			return err
		}

		// Outbox: publish within the same transaction
		txCtx = outbox.WithOutboxMeta(txCtx, "Account", acc.ID)
		s.publishAccountDebited(txCtx, acc, amount)

		result = acc
		return nil
	})
	if err != nil {
		return nil, err
	}

	metrics.WithdrawalsTotal.Inc()
	s.audit(ctx, result.ID, "Debited", map[string]string{
		"amount": fmt.Sprintf("%d", amount), "currency": string(result.Balance.Currency),
	})
	return result, nil
}

// CloseAccount - event sourced close
func (s *Service) CloseAccount(ctx context.Context, accountID string) (err error) {
	defer metrics.ObserveService("AccountService", "CloseAccount", time.Now(), &err)

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		acc, err := s.loadAggregate(txCtx, accountID)
		if err != nil {
			return err
		}
		if err := acc.Close(); err != nil {
			return err
		}
		if err := s.saveAggregate(txCtx, acc); err != nil {
			return err
		}

		// Outbox: publish within the same transaction
		txCtx = outbox.WithOutboxMeta(txCtx, "Account", accountID)
		s.publishAccountClosed(txCtx, accountID)
		return nil
	})
	if err != nil {
		return err
	}

	metrics.AccountsClosedTotal.Inc()
	s.audit(ctx, accountID, "Closed", nil)
	return nil
}

// GetHistory - load all domain events for an account
func (s *Service) GetHistory(ctx context.Context, accountID string) (_ []account.Event, err error) {
	defer metrics.ObserveService("AccountService", "GetHistory", time.Now(), &err)

	events, err := s.eventRepo.LoadEvents(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, account.ErrAccountNotFound
	}
	return events, nil
}

// --- Internal: aggregate load/save ---

func (s *Service) loadAggregate(ctx context.Context, id string) (*account.Account, error) {
	// Pessimistic lock — accounts row'ni qulflash (FOR UPDATE)
	if _, err := s.repo.GetByIDForUpdate(ctx, id); err != nil {
		return nil, err
	}

	snapshot, err := s.eventRepo.LoadSnapshot(ctx, id)
	if err != nil {
		return nil, err
	}

	var acc *account.Account
	fromVersion := 0

	if snapshot != nil {
		acc = account.LoadFromSnapshot(snapshot.State, snapshot.Version, nil)
		fromVersion = snapshot.Version
	}

	events, err := s.eventRepo.LoadEventsFromVersion(ctx, id, fromVersion)
	if err != nil {
		return nil, err
	}

	if acc == nil && len(events) == 0 {
		return nil, account.ErrAccountNotFound
	}

	if acc == nil {
		acc = account.LoadFromHistory(events)
	} else {
		for _, e := range events {
			acc.Apply(e)
		}
	}

	return acc, nil
}

func (s *Service) saveAggregate(ctx context.Context, acc *account.Account) error {
	events := acc.UncommittedEvents()
	if len(events) == 0 {
		return nil
	}

	if err := s.eventRepo.Append(ctx, acc.ID, events); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, acc); err != nil {
		return err
	}

	if acc.Version%snapshotInterval == 0 {
		s.eventRepo.SaveSnapshot(ctx, account.Snapshot{
			AggregateID: acc.ID,
			Version:     acc.Version,
			State:       acc.ToSnapshotState(),
			CreatedAt:   time.Now(),
		})
	}

	acc.ClearUncommittedEvents()
	return nil
}

// --- Audit log helper (async, non-blocking) ---

func (s *Service) audit(ctx context.Context, accountID, action string, attrs map[string]string) {
	if s.auditLog == nil {
		return
	}
	s.auditLog.Log(ctx, domain.AuditEntry{
		AggregateType: "Account",
		AggregateID:   accountID,
		Action:        action,
		Attributes:    attrs,
		Timestamp:     time.Now(),
	})
}

// --- Kafka publish helpers (outbox: called within DB transaction) ---

func newMetadata(userID string) *commonpb.Metadata {
	return &commonpb.Metadata{
		EventId:   uuid.New().String(),
		UserId:    userID,
		Timestamp: timestamppb.Now(),
		Source:    "xbank-api",
	}
}

func (s *Service) publishAccountOpened(ctx context.Context, acc *account.Account, userID string) {
	msg := &accountpb.AccountOpened{
		Metadata:      newMetadata(userID),
		AccountId:     acc.ID,
		UserId:        userID,
		AccountNumber: acc.AccountNumber,
		Currency:      string(acc.Balance.Currency),
	}
	if data, err := proto.Marshal(msg); err == nil {
		s.publisher.Publish(ctx, s.topics.AccountOpened, acc.ID, data)
	}
}

func (s *Service) publishAccountCredited(ctx context.Context, acc *account.Account, amount int64) {
	msg := &accountpb.AccountCredited{
		Metadata:  newMetadata(acc.UserID),
		AccountId: acc.ID,
		Amount:    amount,
		Currency:  string(acc.Balance.Currency),
		Balance:   acc.Balance.Amount,
	}
	if data, err := proto.Marshal(msg); err == nil {
		s.publisher.Publish(ctx, s.topics.AccountCredited, acc.ID, data)
	}
}

func (s *Service) publishAccountDebited(ctx context.Context, acc *account.Account, amount int64) {
	msg := &accountpb.AccountDebited{
		Metadata:  newMetadata(acc.UserID),
		AccountId: acc.ID,
		Amount:    amount,
		Currency:  string(acc.Balance.Currency),
		Balance:   acc.Balance.Amount,
	}
	if data, err := proto.Marshal(msg); err == nil {
		s.publisher.Publish(ctx, s.topics.AccountDebited, acc.ID, data)
	}
}

func (s *Service) publishAccountClosed(ctx context.Context, accountID string) {
	msg := &accountpb.AccountClosed{
		Metadata:  newMetadata(""),
		AccountId: accountID,
	}
	if data, err := proto.Marshal(msg); err == nil {
		s.publisher.Publish(ctx, s.topics.AccountClosed, accountID, data)
	}
}
