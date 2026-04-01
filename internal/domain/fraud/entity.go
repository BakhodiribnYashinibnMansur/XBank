package fraud

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
)

var (
	ErrFraudDetected = apperror.ErrFraudDetected
	ErrAMLBlocked    = apperror.ErrAMLBlocked
)

type RiskLevel string

const (
	RiskLow    RiskLevel = "LOW"    // auto-approve
	RiskMedium RiskLevel = "MEDIUM" // require additional verification
	RiskHigh   RiskLevel = "HIGH"   // block + manual review
)

type Action string

const (
	ActionApprove Action = "APPROVE"
	ActionFlag    Action = "FLAG"
	ActionBlock   Action = "BLOCK"
)

// Check - a fraud/AML check result for a transaction
type Check struct {
	ID            string
	TransferID    string
	UserID        string
	RiskScore     int       // 0-100
	RiskLevel     RiskLevel
	Action        Action
	Flags         []string  // reasons: "LARGE_AMOUNT", "NEW_BENEFICIARY", "HIGH_VELOCITY"
	ReviewedBy    string
	ReviewComment string
	CreatedAt     time.Time
}

// NewCheck - evaluate risk for a transfer
func NewCheck(transferID, userID string, amount int64, flags []string) *Check {
	score := calculateRiskScore(amount, flags)
	level, action := evaluateRisk(score)

	return &Check{
		TransferID: transferID,
		UserID:     userID,
		RiskScore:  score,
		RiskLevel:  level,
		Action:     action,
		Flags:      flags,
		CreatedAt:  time.Now(),
	}
}

// calculateRiskScore - weighted scoring (0-100)
func calculateRiskScore(amount int64, flags []string) int {
	score := 0

	// Amount-based scoring
	if amount > 1000000000 { // > 10,000,000 UZS (10M)
		score += 40
	} else if amount > 100000000 { // > 1,000,000 UZS (1M)
		score += 20
	} else if amount > 10000000 { // > 100,000 UZS
		score += 10
	}

	// Flag-based scoring
	flagScores := map[string]int{
		"LARGE_AMOUNT":    15,
		"NEW_BENEFICIARY": 10,
		"HIGH_VELOCITY":   20,
		"UNUSUAL_TIME":    10,
		"NEW_DEVICE":      15,
		"ROUND_AMOUNT":    5,
	}

	for _, f := range flags {
		if s, ok := flagScores[f]; ok {
			score += s
		}
	}

	if score > 100 {
		score = 100
	}
	return score
}

func evaluateRisk(score int) (RiskLevel, Action) {
	switch {
	case score >= 70:
		return RiskHigh, ActionBlock
	case score >= 40:
		return RiskMedium, ActionFlag
	default:
		return RiskLow, ActionApprove
	}
}

type Repository interface {
	Create(ctx context.Context, check *Check) error
	GetByTransferID(ctx context.Context, transferID string) (*Check, error)
	ListFlagged(ctx context.Context, limit, offset int) ([]*Check, error)
	CountFlagged(ctx context.Context) (int64, error)
	Update(ctx context.Context, check *Check) error
}
