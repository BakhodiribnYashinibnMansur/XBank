package http

import domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/ops/generic/ratelimit/domain"

func toResponse(r *domain.RateLimitRule) RateLimitResponse {
	return RateLimitResponse{
		ID:            r.ID,
		Key:           r.Key,
		MaxRequests:   r.MaxRequests,
		WindowSeconds: r.WindowSeconds,
		Description:   r.Description,
		Enabled:       r.Enabled,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}
}
