package clockskew

import (
	"time"
)

// Result describes the outcome of a clock skew check.
type Result struct {
	Drift    time.Duration // positive = client ahead, negative = client behind
	Accepted bool          // true if drift is within the allowed skew
}

// Detector checks whether a client timestamp is within acceptable skew
// relative to the server clock.
type Detector struct {
	maxSkew time.Duration
}

// NewDetector creates a Detector with the given maximum allowed skew.
func NewDetector(maxSkew time.Duration) *Detector {
	return &Detector{maxSkew: maxSkew}
}

// Check validates a Unix timestamp (seconds) against the server clock.
func (d *Detector) Check(clientTimestamp int64) Result {
	return d.CheckTime(time.Unix(clientTimestamp, 0))
}

// CheckTime validates a time.Time against the server clock.
func (d *Detector) CheckTime(clientTime time.Time) Result {
	drift := clientTime.Sub(time.Now())
	accepted := drift >= -d.maxSkew && drift <= d.maxSkew
	return Result{
		Drift:    drift,
		Accepted: accepted,
	}
}

// MaxSkew returns the configured maximum allowed skew.
func (d *Detector) MaxSkew() time.Duration {
	return d.maxSkew
}
