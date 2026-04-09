package domain

import (
	"testing"
)

func TestNewCheck_LowRisk(t *testing.T) {
	check := NewCheck("tx-1", "user-1", 5000000, nil) // 50,000 UZS, no flags
	if check.RiskLevel != RiskLow {
		t.Errorf("expected LOW risk, got: %s", check.RiskLevel)
	}
	if check.Action != ActionApprove {
		t.Errorf("expected APPROVE action, got: %s", check.Action)
	}
	if check.TransferID != "tx-1" {
		t.Errorf("TransferID expected tx-1, got: %s", check.TransferID)
	}
	if check.UserID != "user-1" {
		t.Errorf("UserID expected user-1, got: %s", check.UserID)
	}
}

func TestNewCheck_MediumRisk(t *testing.T) {
	// Large amount + new beneficiary should give medium risk
	flags := []string{"LARGE_AMOUNT", "NEW_BENEFICIARY"}
	check := NewCheck("tx-2", "user-1", 200000000, flags) // 2M UZS + flags

	if check.RiskLevel != RiskMedium {
		t.Errorf("expected MEDIUM risk, got: %s (score: %d)", check.RiskLevel, check.RiskScore)
	}
	if check.Action != ActionFlag {
		t.Errorf("expected FLAG action, got: %s", check.Action)
	}
}

func TestNewCheck_HighRisk(t *testing.T) {
	// Very large amount + multiple flags should give high risk
	flags := []string{"LARGE_AMOUNT", "HIGH_VELOCITY", "NEW_DEVICE", "UNUSUAL_TIME"}
	check := NewCheck("tx-3", "user-1", 2000000000, flags) // 20M UZS + many flags

	if check.RiskLevel != RiskHigh {
		t.Errorf("expected HIGH risk, got: %s (score: %d)", check.RiskLevel, check.RiskScore)
	}
	if check.Action != ActionBlock {
		t.Errorf("expected BLOCK action, got: %s", check.Action)
	}
}

func TestCalculateRiskScore_AmountThresholds(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		minScore int
	}{
		{"small amount", 1000000, 0},        // 10,000 UZS
		{"medium amount", 50000000, 10},      // 500,000 UZS (> 100K threshold)
		{"large amount", 500000000, 20},      // 5M UZS (> 1M threshold)
		{"very large amount", 2000000000, 40}, // 20M UZS (> 10M threshold)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateRiskScore(tt.amount, nil)
			if score < tt.minScore {
				t.Errorf("score for amount %d expected >= %d, got: %d", tt.amount, tt.minScore, score)
			}
		})
	}
}

func TestCalculateRiskScore_Flags(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		want  int
	}{
		{"no flags", nil, 0},
		{"single flag", []string{"LARGE_AMOUNT"}, 15},
		{"unknown flag ignored", []string{"UNKNOWN_FLAG"}, 0},
		{"multiple flags", []string{"LARGE_AMOUNT", "NEW_BENEFICIARY"}, 25},
		{"high velocity", []string{"HIGH_VELOCITY"}, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateRiskScore(0, tt.flags)
			if score != tt.want {
				t.Errorf("score expected %d, got: %d", tt.want, score)
			}
		})
	}
}

func TestCalculateRiskScore_MaxCap(t *testing.T) {
	// Score should be capped at 100
	flags := []string{"LARGE_AMOUNT", "NEW_BENEFICIARY", "HIGH_VELOCITY", "UNUSUAL_TIME", "NEW_DEVICE", "ROUND_AMOUNT"}
	score := calculateRiskScore(2000000000, flags) // max amount + all flags
	if score > 100 {
		t.Errorf("score should be capped at 100, got: %d", score)
	}
}

func TestEvaluateRisk(t *testing.T) {
	tests := []struct {
		score     int
		wantLevel RiskLevel
		wantAction Action
	}{
		{0, RiskLow, ActionApprove},
		{39, RiskLow, ActionApprove},
		{40, RiskMedium, ActionFlag},
		{69, RiskMedium, ActionFlag},
		{70, RiskHigh, ActionBlock},
		{100, RiskHigh, ActionBlock},
	}

	for _, tt := range tests {
		level, action := evaluateRisk(tt.score)
		if level != tt.wantLevel {
			t.Errorf("score %d: level expected %s, got: %s", tt.score, tt.wantLevel, level)
		}
		if action != tt.wantAction {
			t.Errorf("score %d: action expected %s, got: %s", tt.score, tt.wantAction, action)
		}
	}
}
