package http

import "time"

type CreateIPRuleRequest struct {
	IPAddress string     `json:"ip_address"`
	RuleType  string     `json:"rule_type"` // ALLOW or DENY
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type IPRuleResponse struct {
	ID        string     `json:"id"`
	IPAddress string     `json:"ip_address"`
	RuleType  string     `json:"rule_type"`
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedBy string     `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
}
