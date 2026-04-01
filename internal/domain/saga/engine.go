package saga

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// Status - saga execution status
type Status string

const (
	StatusRunning      Status = "RUNNING"
	StatusCompleted    Status = "COMPLETED"
	StatusCompensating Status = "COMPENSATING"
	StatusFailed       Status = "FAILED"
)

// Step - a single saga step with execute + compensate
type Step struct {
	Name       string
	Execute    func(ctx context.Context) error
	Compensate func(ctx context.Context) error // nil = no compensation needed
}

// Result - saga execution result
type Result struct {
	SagaID      string
	Status      Status
	CompletedAt time.Time
	FailedStep  string
	Error       error
	StepsRun    int
}

// Execute - run all steps in order, compensate on failure
func Execute(ctx context.Context, sagaID string, steps []Step) *Result {
	result := &Result{
		SagaID: sagaID,
		Status: StatusRunning,
	}

	logger.Log.Info("saga started",
		zap.String("saga_id", sagaID),
		zap.Int("steps", len(steps)),
	)

	// Execute steps forward
	for i, step := range steps {
		logger.Log.Debug("saga step executing",
			zap.String("saga_id", sagaID),
			zap.String("step", step.Name),
			zap.Int("index", i+1),
		)

		if err := step.Execute(ctx); err != nil {
			result.Status = StatusCompensating
			result.FailedStep = step.Name
			result.Error = err

			logger.Log.Error("saga step failed, compensating",
				zap.String("saga_id", sagaID),
				zap.String("step", step.Name),
				zap.Error(err),
			)

			// Compensate in reverse order (skip the failed step)
			compensate(ctx, sagaID, steps[:i])

			result.Status = StatusFailed
			result.CompletedAt = time.Now()
			result.StepsRun = i
			return result
		}
	}

	result.Status = StatusCompleted
	result.CompletedAt = time.Now()
	result.StepsRun = len(steps)

	logger.Log.Info("saga completed",
		zap.String("saga_id", sagaID),
		zap.Int("steps_run", result.StepsRun),
	)

	return result
}

// compensate - run compensations in reverse order
func compensate(ctx context.Context, sagaID string, completedSteps []Step) {
	for i := len(completedSteps) - 1; i >= 0; i-- {
		step := completedSteps[i]
		if step.Compensate == nil {
			continue
		}

		logger.Log.Info("saga compensating",
			zap.String("saga_id", sagaID),
			zap.String("step", step.Name),
		)

		if err := step.Compensate(ctx); err != nil {
			// Compensation failure is critical — log and continue
			logger.Log.Error("saga compensation failed",
				zap.String("saga_id", sagaID),
				zap.String("step", step.Name),
				zap.Error(err),
			)
		}
	}
}

// GenerateID - create a unique saga ID
func GenerateID() string {
	return fmt.Sprintf("saga-%d", time.Now().UnixNano())
}
