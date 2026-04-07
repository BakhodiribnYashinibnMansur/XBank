package circuitbreaker

import (
	"errors"
	"sync"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

var ErrCircuitOpen = errors.New("circuit breaker is open, service unavailable")

// State - circuit breaker state
type State int

const (
	Closed   State = iota // normal operation
	Open                  // failing, reject all calls
	HalfOpen              // testing if service recovered
)

// Breaker - circuit breaker for external service calls
type Breaker struct {
	mu              sync.RWMutex
	name            string
	state           State
	failures        int
	maxFailures     int
	resetTimeout    time.Duration
	lastFailureTime time.Time
}

func New(name string, maxFailures int, resetTimeout time.Duration) *Breaker {
	return &Breaker{
		name:         name,
		state:        Closed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// Execute - run a function through the circuit breaker
func (b *Breaker) Execute(fn func() error) error {
	if !b.allowRequest() {
		return ErrCircuitOpen
	}

	err := fn()
	b.recordResult(err)
	return err
}

func (b *Breaker) allowRequest() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	switch b.state {
	case Closed:
		return true
	case Open:
		// Check if reset timeout passed → move to half-open
		if time.Since(b.lastFailureTime) > b.resetTimeout {
			b.mu.RUnlock()
			b.mu.Lock()
			b.state = HalfOpen
			b.mu.Unlock()
			b.mu.RLock()
			return true
		}
		return false
	case HalfOpen:
		return true // allow one test request
	}
	return true
}

func (b *Breaker) recordResult(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err == nil {
		// Success
		if b.state == HalfOpen {
			b.state = Closed
			b.failures = 0
			logger.Log.Info("circuit breaker closed (recovered)", zap.String("service", b.name))
		}
		b.failures = 0
		return
	}

	// Failure
	b.failures++
	b.lastFailureTime = time.Now()

	if b.failures >= b.maxFailures {
		b.state = Open
		logger.Log.Warn("circuit breaker opened",
			zap.String("service", b.name),
			zap.Int("failures", b.failures),
			zap.Duration("reset_timeout", b.resetTimeout),
		)
	}
}

// State - current state
func (b *Breaker) GetState() State {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.state
}
