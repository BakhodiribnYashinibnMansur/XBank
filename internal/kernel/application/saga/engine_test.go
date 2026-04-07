package saga

import (
	"context"
	"errors"
	"testing"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
)

func init() {
	logger.Init(true)
}

func TestExecute_AllStepsPass(t *testing.T) {
	steps := []Step{
		{Name: "step1", Execute: func(ctx context.Context) error { return nil }},
		{Name: "step2", Execute: func(ctx context.Context) error { return nil }},
		{Name: "step3", Execute: func(ctx context.Context) error { return nil }},
	}

	result := Execute(context.Background(), "test-saga", steps)

	if result.Status != StatusCompleted {
		t.Errorf("expected COMPLETED, got: %s", result.Status)
	}
	if result.StepsRun != 3 {
		t.Errorf("expected 3 steps, got: %d", result.StepsRun)
	}
}

func TestExecute_FailAtStep2_Compensates(t *testing.T) {
	compensated := false

	steps := []Step{
		{
			Name:       "step1",
			Execute:    func(ctx context.Context) error { return nil },
			Compensate: func(ctx context.Context) error { compensated = true; return nil },
		},
		{
			Name:    "step2_fail",
			Execute: func(ctx context.Context) error { return errors.New("boom") },
		},
		{
			Name:    "step3_never",
			Execute: func(ctx context.Context) error { t.Fatal("should not run"); return nil },
		},
	}

	result := Execute(context.Background(), "test-saga", steps)

	if result.Status != StatusFailed {
		t.Errorf("expected FAILED, got: %s", result.Status)
	}
	if result.FailedStep != "step2_fail" {
		t.Errorf("expected failed step: step2_fail, got: %s", result.FailedStep)
	}
	if !compensated {
		t.Error("step1 should have been compensated")
	}
	if result.StepsRun != 1 {
		t.Errorf("expected 1 step run, got: %d", result.StepsRun)
	}
}

func TestExecute_FailAtStep1_NoCompensation(t *testing.T) {
	steps := []Step{
		{
			Name:    "step1_fail",
			Execute: func(ctx context.Context) error { return errors.New("boom") },
		},
	}

	result := Execute(context.Background(), "test-saga", steps)

	if result.Status != StatusFailed {
		t.Errorf("expected FAILED, got: %s", result.Status)
	}
	if result.StepsRun != 0 {
		t.Errorf("expected 0 steps, got: %d", result.StepsRun)
	}
}
