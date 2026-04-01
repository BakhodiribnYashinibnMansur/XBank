package saga

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/account"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/fraud"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/ledger"
	domainSaga "github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/saga"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/transfer"
)

// TransferSaga - orchestrates a fund transfer through multiple steps
type TransferSaga struct {
	accountRepo  account.Repository
	transferRepo transfer.Repository
	fraudRepo    fraud.Repository
	ledgerRepo   ledger.Repository
	txManager    shared.TxManager
}

func NewTransferSaga(
	accountRepo account.Repository,
	transferRepo transfer.Repository,
	fraudRepo fraud.Repository,
	ledgerRepo ledger.Repository,
	txManager shared.TxManager,
) *TransferSaga {
	return &TransferSaga{
		accountRepo:  accountRepo,
		transferRepo: transferRepo,
		fraudRepo:    fraudRepo,
		ledgerRepo:   ledgerRepo,
		txManager:    txManager,
	}
}

// Execute - run the transfer saga
func (s *TransferSaga) Execute(ctx context.Context, tr *transfer.Transfer) *domainSaga.Result {
	sagaID := domainSaga.GenerateID()
	amount := tr.Amount.Amount
	currency := string(tr.Amount.Currency)

	var fromAcc, toAcc *account.Account

	steps := []domainSaga.Step{
		{
			Name: "validate_accounts",
			Execute: func(ctx context.Context) error {
				var err error
				fromAcc, err = s.accountRepo.GetByID(ctx, tr.FromAccountID)
				if err != nil {
					return err
				}
				toAcc, err = s.accountRepo.GetByID(ctx, tr.ToAccountID)
				if err != nil {
					return err
				}
				if fromAcc.Balance.Currency != tr.Amount.Currency || toAcc.Balance.Currency != tr.Amount.Currency {
					return shared.ErrCurrencyMismatch
				}
				if fromAcc.Balance.Amount < amount {
					return shared.ErrInsufficientFunds
				}
				return nil
			},
		},
		{
			Name: "fraud_check",
			Execute: func(ctx context.Context) error {
				check := fraud.NewCheck(tr.ID, "", amount, nil)
				if err := s.fraudRepo.Create(ctx, check); err != nil {
					return err
				}
				if check.Action == fraud.ActionBlock {
					return fraud.ErrAMLBlocked
				}
				return nil
			},
		},
		{
			Name: "debit_sender",
			Execute: func(ctx context.Context) error {
				return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
					acc, err := s.accountRepo.GetByIDForUpdate(txCtx, tr.FromAccountID)
					if err != nil {
						return err
					}
					money, _ := shared.NewMoney(amount, acc.Balance.Currency)
					if err := acc.Withdraw(money); err != nil {
						return err
					}
					return s.accountRepo.Update(txCtx, acc)
				})
			},
			Compensate: func(ctx context.Context) error {
				// Reverse: credit back to sender
				return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
					acc, err := s.accountRepo.GetByIDForUpdate(txCtx, tr.FromAccountID)
					if err != nil {
						return err
					}
					money, _ := shared.NewMoney(amount, acc.Balance.Currency)
					acc.Deposit(money)
					return s.accountRepo.Update(txCtx, acc)
				})
			},
		},
		{
			Name: "credit_receiver",
			Execute: func(ctx context.Context) error {
				return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
					acc, err := s.accountRepo.GetByIDForUpdate(txCtx, tr.ToAccountID)
					if err != nil {
						return err
					}
					money, _ := shared.NewMoney(amount, acc.Balance.Currency)
					if err := acc.Deposit(money); err != nil {
						return err
					}
					return s.accountRepo.Update(txCtx, acc)
				})
			},
			Compensate: func(ctx context.Context) error {
				// Reverse: debit from receiver
				return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
					acc, err := s.accountRepo.GetByIDForUpdate(txCtx, tr.ToAccountID)
					if err != nil {
						return err
					}
					money, _ := shared.NewMoney(amount, acc.Balance.Currency)
					acc.Withdraw(money)
					return s.accountRepo.Update(txCtx, acc)
				})
			},
		},
		{
			Name: "record_ledger",
			Execute: func(ctx context.Context) error {
				debit, credit := ledger.CreatePair(tr.ID, tr.FromAccountID, tr.ToAccountID, amount, currency)
				return s.ledgerRepo.CreatePair(ctx, debit, credit)
			},
		},
		{
			Name: "complete_transfer",
			Execute: func(ctx context.Context) error {
				tr.Complete()
				return s.transferRepo.Update(ctx, tr)
			},
		},
	}

	return domainSaga.Execute(ctx, sagaID, steps)
}
