package http

import domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/supporting/iprule/domain"

func toResponse(r *domain.IPRule) IPRuleResponse {
	return IPRuleResponse{
		ID:        r.ID,
		IPAddress: r.IPAddress,
		RuleType:  string(r.RuleType),
		Reason:    r.Reason,
		ExpiresAt: r.ExpiresAt,
		CreatedBy: r.CreatedBy,
		CreatedAt: r.CreatedAt,
	}
}
