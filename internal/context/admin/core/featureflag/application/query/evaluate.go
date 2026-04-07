package query

import (
	"context"
	"sort"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/featureflag/domain"
)

// EvaluateHandler evaluates a feature flag for a specific user/context.
// Priority: 1) Rule groups (sorted by priority) 2) Rollout % 3) Default value
type EvaluateHandler struct {
	writeRepo domain.WriteRepository
}

func NewEvaluateHandler(writeRepo domain.WriteRepository) *EvaluateHandler {
	return &EvaluateHandler{writeRepo: writeRepo}
}

func (h *EvaluateHandler) Handle(ctx context.Context, req application.EvaluateRequest) (*application.EvaluateResponse, error) {
	flag, err := h.writeRepo.FindByKey(ctx, req.Key)
	if err != nil || flag == nil {
		return &application.EvaluateResponse{
			Key:     req.Key,
			Value:   "",
			Enabled: false,
		}, domain.ErrFlagNotFound
	}

	// Inactive → always default
	if !flag.Active {
		return &application.EvaluateResponse{
			Key:     req.Key,
			Value:   flag.DefaultValue,
			Enabled: false,
		}, nil
	}

	// Check rule groups (sorted by priority, lower = first)
	if len(flag.RuleGroups) > 0 {
		sorted := make([]domain.RuleGroup, len(flag.RuleGroups))
		copy(sorted, flag.RuleGroups)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Priority < sorted[j].Priority
		})

		for _, rg := range sorted {
			if rg.Matches(req.Attributes) {
				return &application.EvaluateResponse{
					Key:     req.Key,
					Value:   rg.Value,
					Enabled: true,
				}, nil
			}
		}
	}

	// Rollout percentage check
	enabled := flag.IsEnabledForUser(req.UserID)
	value := flag.DefaultValue
	if !enabled {
		value = flag.DefaultValue
	}

	return &application.EvaluateResponse{
		Key:     req.Key,
		Value:   value,
		Enabled: enabled,
	}, nil
}
