package account

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/account"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/config"
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
	txManager shared.TxManager
	publisher shared.EventPublisher
	topics    config.KafkaTopicsConfig
}

func NewService(
	repo account.Repository,
	eventRepo account.EventRepository,
	txManager shared.TxManager,
	publisher shared.EventPublisher,
	topics config.KafkaTopicsConfig,
) *Service {
	return &Service{
		repo:      repo,
		eventRepo: eventRepo,
		txManager: txManager,
		publisher: publisher,
		topics:    topics,
	}
}

// CreateAccount - open a new account (event sourced)
func (s *Service) CreateAccount(ctx context.Context, userID string, currency shared.Currency) (*account.Account, error) {
	acc, err := account.NewAccount(userID, currency)
	if err != nil {
		return nil, err
	}

	err = s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		if err := s.eventRepo.Append(txCtx, acc.ID, 0, acc.UncommittedEvents()); err != nil {
			return err
		}
		if err := s.repo.Create(txCtx, acc); err != nil {
			return err
		}
		acc.ClearUncommittedEvents()
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.publishAccountOpened(ctx, acc, userID)
	return acc, nil
}

// GetByID - read from projection (fast)
func (s *Service) GetByID(ctx context.Context, id string) (*account.Account, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByUserID - read from projection (paginated)
func (s *Service) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*account.Account, int64, error) {
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
func (s *Service) Deposit(ctx context.Context, accountID string, amount int64) (*account.Account, error) {
	var result *account.Account

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		acc, err := s.loadAggregate(txCtx, accountID)
		if err != nil {
			return err
		}

		money, err := shared.NewMoney(amount, acc.Balance.Currency)
		if err != nil {
			return err
		}
		if err := acc.Deposit(money); err != nil {
			return err
		}

		if err := s.saveAggregate(txCtx, acc); err != nil {
			return err
		}
		result = acc
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.publishAccountCredited(ctx, result, amount)
	return result, nil
}

// Withdraw - event sourced withdrawal
func (s *Service) Withdraw(ctx context.Context, accountID string, amount int64) (*account.Account, error) {
	var result *account.Account

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		acc, err := s.loadAggregate(txCtx, accountID)
		if err != nil {
			return err
		}

		money, err := shared.NewMoney(amount, acc.Balance.Currency)
		if err != nil {
			return err
		}
		if err := acc.Withdraw(money); err != nil {
			return err
		}

		if err := s.saveAggregate(txCtx, acc); err != nil {
			return err
		}
		result = acc
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.publishAccountDebited(ctx, result, amount)
	return result, nil
}

// CloseAccount - event sourced close
func (s *Service) CloseAccount(ctx context.Context, accountID string) error {
	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		acc, err := s.loadAggregate(txCtx, accountID)
		if err != nil {
			return err
		}
		if err := acc.Close(); err != nil {
			return err
		}
		return s.saveAggregate(txCtx, acc)
	})
	if err != nil {
		return err
	}

	s.publishAccountClosed(ctx, accountID)
	return nil
}

// GetHistory - load all domain events for an account
func (s *Service) GetHistory(ctx context.Context, accountID string) ([]account.Event, error) {
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

	expectedVersion := acc.Version - len(events)

	if err := s.eventRepo.Append(ctx, acc.ID, expectedVersion, events); err != nil {
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

// --- Kafka publish helpers (best-effort, after DB commit) ---

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
