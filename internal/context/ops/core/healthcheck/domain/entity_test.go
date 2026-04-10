package domain

import (
	"testing"
	"time"
)

func TestNewSystemHealth(t *testing.T) {
	tests := []struct {
		name       string
		checks     []ComponentCheck
		wantStatus ComponentStatus
	}{
		{
			"all healthy",
			[]ComponentCheck{
				{Name: "postgres", Status: StatusHealthy, Latency: 5 * time.Millisecond},
				{Name: "redis", Status: StatusHealthy, Latency: 2 * time.Millisecond},
			},
			StatusHealthy,
		},
		{
			"one degraded",
			[]ComponentCheck{
				{Name: "postgres", Status: StatusHealthy},
				{Name: "redis", Status: StatusDegraded, Message: "slow response"},
			},
			StatusDegraded,
		},
		{
			"one unhealthy overrides degraded",
			[]ComponentCheck{
				{Name: "postgres", Status: StatusUnhealthy, Message: "connection refused"},
				{Name: "redis", Status: StatusDegraded},
				{Name: "kafka", Status: StatusHealthy},
			},
			StatusUnhealthy,
		},
		{
			"empty components",
			[]ComponentCheck{},
			StatusHealthy,
		},
		{
			"nil components",
			nil,
			StatusHealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			health := NewSystemHealth(tt.checks)
			if health.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", health.Status, tt.wantStatus)
			}
			if len(health.Components) != len(tt.checks) {
				t.Errorf("components count = %d, want %d", len(health.Components), len(tt.checks))
			}
			if health.CheckedAt.IsZero() {
				t.Error("checked_at should not be zero")
			}
		})
	}
}

func TestComponentStatusConstants(t *testing.T) {
	if StatusHealthy != "HEALTHY" {
		t.Errorf("StatusHealthy = %q", StatusHealthy)
	}
	if StatusDegraded != "DEGRADED" {
		t.Errorf("StatusDegraded = %q", StatusDegraded)
	}
	if StatusUnhealthy != "UNHEALTHY" {
		t.Errorf("StatusUnhealthy = %q", StatusUnhealthy)
	}
}
