package command

import (
	"context"
	"time"

	audit "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/audit/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

type Service struct {
	repo audit.Repository
}

func NewService(repo audit.Repository) *Service {
	return &Service{repo: repo}
}

// CreateAuditLogInput holds the input for creating an audit log entry.
type CreateAuditLogInput struct {
	AggregateType string
	AggregateID   string
	Action        string
	ActorID       string
	Attributes    map[string]any
	IPAddress     string
	UserAgent     string
}

func (s *Service) CreateAuditLog(ctx context.Context, input CreateAuditLogInput) (_ *audit.AuditLog, err error) {
	defer metrics.ObserveService("AuditService", "CreateAuditLog", time.Now(), &err)

	log, err := audit.NewAuditLog(
		input.AggregateType, input.AggregateID, input.Action,
		input.ActorID, input.Attributes, input.IPAddress, input.UserAgent,
	)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateAuditLog(ctx, log); err != nil {
		return nil, err
	}
	return log, nil
}

func (s *Service) ListAuditLogs(ctx context.Context, filter audit.AuditFilter) (_ []*audit.AuditLog, _ int64, err error) {
	defer metrics.ObserveService("AuditService", "ListAuditLogs", time.Now(), &err)
	return s.repo.ListAuditLogs(ctx, filter)
}

// CreateEndpointHistoryInput holds the input for creating an endpoint history entry.
type CreateEndpointHistoryInput struct {
	Method     string
	Path       string
	StatusCode int
	UserID     string
	IPAddress  string
	DurationMs int
}

func (s *Service) CreateEndpointHistory(ctx context.Context, input CreateEndpointHistoryInput) (_ *audit.EndpointHistory, err error) {
	defer metrics.ObserveService("AuditService", "CreateEndpointHistory", time.Now(), &err)

	h, err := audit.NewEndpointHistory(
		input.Method, input.Path, input.StatusCode,
		input.UserID, input.IPAddress, input.DurationMs,
	)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateEndpointHistory(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

func (s *Service) ListEndpointHistory(ctx context.Context, filter audit.EndpointFilter) (_ []*audit.EndpointHistory, _ int64, err error) {
	defer metrics.ObserveService("AuditService", "ListEndpointHistory", time.Now(), &err)
	return s.repo.ListEndpointHistory(ctx, filter)
}
