package domain

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

var (
	ErrIPRuleNotFound = domain.NewDomainError("IP_RULE_NOT_FOUND", "IP rule not found")
	ErrIPRuleExists   = domain.NewDomainError("IP_RULE_EXISTS", "IP rule for this address already exists")
)

type RuleType string

const (
	RuleTypeAllow RuleType = "ALLOW"
	RuleTypeDeny  RuleType = "DENY"
)

// IPRule defines an IP-based access control rule.
type IPRule struct {
	ID        string
	IPAddress string
	RuleType  RuleType
	Reason    string
	ExpiresAt *time.Time
	CreatedBy string
	CreatedAt time.Time
}

// NewIPRule creates a new IP rule.
func NewIPRule(ipAddress string, ruleType RuleType, reason, createdBy string, expiresAt *time.Time) (*IPRule, error) {
	if ipAddress == "" {
		return nil, domain.NewDomainError("MISSING_FIELD", "ip_address is required")
	}
	if ruleType != RuleTypeAllow && ruleType != RuleTypeDeny {
		return nil, domain.NewDomainError("INVALID_FIELD", "rule_type must be ALLOW or DENY")
	}

	return &IPRule{
		IPAddress: ipAddress,
		RuleType:  ruleType,
		Reason:    reason,
		ExpiresAt: expiresAt,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}, nil
}

// IsExpired returns true if the rule has an expiration and it has passed.
func (r *IPRule) IsExpired() bool {
	if r.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*r.ExpiresAt)
}

// Repository defines the persistence contract for IP rules.
type Repository interface {
	Save(ctx context.Context, rule *IPRule) error
	FindByID(ctx context.Context, id string) (*IPRule, error)
	FindByIP(ctx context.Context, ipAddress string) (*IPRule, error)
	ListAll(ctx context.Context) ([]*IPRule, error)
	Delete(ctx context.Context, id string) error
}
