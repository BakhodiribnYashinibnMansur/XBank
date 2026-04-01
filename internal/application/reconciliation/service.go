package reconciliation

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/account"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/ledger"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// Result - reconciliation check result for one account
type Result struct {
	AccountID      string `json:"account_id"`
	ProjectedBalance int64  `json:"projected_balance"` // from accounts table
	LedgerBalance  int64  `json:"ledger_balance"`    // from ledger entries
	Match          bool   `json:"match"`
}

// Service - daily reconciliation checks
type Service struct {
	accountRepo account.Repository
	ledgerRepo  ledger.Repository
}

func NewService(accountRepo account.Repository, ledgerRepo ledger.Repository) *Service {
	return &Service{accountRepo: accountRepo, ledgerRepo: ledgerRepo}
}

// CheckAccount - verify account balance matches ledger sum
func (s *Service) CheckAccount(ctx context.Context, accountID string) (*Result, error) {
	acc, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	ledgerBalance, err := s.ledgerRepo.BalanceByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	result := &Result{
		AccountID:        accountID,
		ProjectedBalance: acc.Balance.Amount,
		LedgerBalance:    ledgerBalance,
		Match:            acc.Balance.Amount == ledgerBalance,
	}

	if !result.Match {
		logger.Log.Error("reconciliation mismatch",
			zap.String("account_id", accountID),
			zap.Int64("projected", result.ProjectedBalance),
			zap.Int64("ledger", result.LedgerBalance),
			zap.Int64("diff", result.ProjectedBalance-result.LedgerBalance),
		)
	}

	return result, nil
}

// CheckAllAccounts - run reconciliation for all accounts of a user
func (s *Service) CheckAllAccounts(ctx context.Context, userID string) ([]*Result, error) {
	accounts, err := s.accountRepo.ListByUserID(ctx, userID, 1000, 0)
	if err != nil {
		return nil, err
	}

	var results []*Result
	for _, acc := range accounts {
		result, err := s.CheckAccount(ctx, acc.ID)
		if err != nil {
			logger.Log.Error("reconciliation check failed",
				zap.String("account_id", acc.ID),
				zap.Error(err),
			)
			results = append(results, &Result{
				AccountID: acc.ID,
				Match:     false,
			})
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// Summary - count matches and mismatches
func Summary(results []*Result) string {
	matches, mismatches := 0, 0
	for _, r := range results {
		if r.Match {
			matches++
		} else {
			mismatches++
		}
	}
	return fmt.Sprintf("total=%d matches=%d mismatches=%d", len(results), matches, mismatches)
}
