package latency

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Tracker records operation latency using a Prometheus histogram.
type Tracker struct {
	histogram *prometheus.HistogramVec
}

// New creates a latency tracker with the given metric name and labels.
func New(namespace, subsystem, name, help string, labels []string) *Tracker {
	h := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      name,
			Help:      help,
			Buckets:   prometheus.DefBuckets,
		},
		labels,
	)
	prometheus.MustRegister(h)
	return &Tracker{histogram: h}
}

// Observe records a duration observation with the given label values.
func (t *Tracker) Observe(duration time.Duration, labelValues ...string) {
	t.histogram.WithLabelValues(labelValues...).Observe(duration.Seconds())
}

// Since records the time elapsed since start with the given label values.
func (t *Tracker) Since(start time.Time, labelValues ...string) {
	t.Observe(time.Since(start), labelValues...)
}

// Timer returns a function that, when called, records the elapsed time.
// Usage:
//
//	done := tracker.Timer("operation")
//	defer done()
func (t *Tracker) Timer(labelValues ...string) func() {
	start := time.Now()
	return func() {
		t.Since(start, labelValues...)
	}
}
