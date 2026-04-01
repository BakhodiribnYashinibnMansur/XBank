// Package metrics defines all Prometheus metrics for the XBank application.
//
// Metric naming: xbank_{subsystem}_{name}_{unit}
// Labels follow Prometheus best practices.
package metrics

import "github.com/prometheus/client_golang/prometheus"

// ── HTTP Metrics ─────────────────────────────────────

var (
	// HTTPRequestsTotal — total number of HTTP requests by method, path, status.
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "xbank",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration — request latency histogram in seconds.
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "xbank",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"method", "path"},
	)

	// HTTPActiveRequests — number of in-flight requests.
	HTTPActiveRequests = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "xbank",
			Subsystem: "http",
			Name:      "active_requests",
			Help:      "Number of active HTTP requests being processed",
		},
	)
)

// ── Database Metrics ─────────────────────────────────

var (
	// DBQueryDuration — SQL query duration histogram in seconds.
	DBQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "xbank",
			Subsystem: "db",
			Name:      "query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5},
		},
		[]string{"operation"},
	)

	// DBQueryTotal — total number of database queries.
	DBQueryTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "xbank",
			Subsystem: "db",
			Name:      "queries_total",
			Help:      "Total number of database queries",
		},
		[]string{"operation", "status"},
	)

	// DBSlowQueries — total number of queries exceeding the slow threshold.
	DBSlowQueries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "xbank",
			Subsystem: "db",
			Name:      "slow_queries_total",
			Help:      "Total number of slow database queries (>100ms)",
		},
		[]string{"operation"},
	)

	// DBPoolSize — current number of connections in the pool.
	DBPoolSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "xbank",
			Subsystem: "db",
			Name:      "pool_connections",
			Help:      "Number of database pool connections by state",
		},
		[]string{"state"},
	)
)

// ── Kafka Metrics ────────────────────────────────────

var (
	// KafkaMessagesTotal — total number of Kafka messages published.
	KafkaMessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "xbank",
			Subsystem: "kafka",
			Name:      "messages_total",
			Help:      "Total number of Kafka messages published",
		},
		[]string{"topic", "status"},
	)

	// KafkaPublishDuration — Kafka publish latency in seconds.
	KafkaPublishDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "xbank",
			Subsystem: "kafka",
			Name:      "publish_duration_seconds",
			Help:      "Kafka message publish duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"topic"},
	)
)

// ── Business Metrics ─────────────────────────────────

var (
	// AccountsCreatedTotal — total accounts created.
	AccountsCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "xbank",
			Subsystem: "business",
			Name:      "accounts_created_total",
			Help:      "Total number of accounts created",
		},
	)

	// AccountsClosedTotal — total accounts closed.
	AccountsClosedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "xbank",
			Subsystem: "business",
			Name:      "accounts_closed_total",
			Help:      "Total number of accounts closed",
		},
	)

	// TransfersTotal — total transfers by status.
	TransfersTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "xbank",
			Subsystem: "business",
			Name:      "transfers_total",
			Help:      "Total number of transfers by status",
		},
		[]string{"status"},
	)

	// DepositsTotal — total deposit operations.
	DepositsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "xbank",
			Subsystem: "business",
			Name:      "deposits_total",
			Help:      "Total number of deposit operations",
		},
	)

	// WithdrawalsTotal — total withdrawal operations.
	WithdrawalsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "xbank",
			Subsystem: "business",
			Name:      "withdrawals_total",
			Help:      "Total number of withdrawal operations",
		},
	)

	// LoginsTotal — total login attempts by result.
	LoginsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "xbank",
			Subsystem: "business",
			Name:      "logins_total",
			Help:      "Total number of login attempts",
		},
		[]string{"status"},
	)
)

// Register registers all custom metrics with the default Prometheus registry.
func Register() {
	// HTTP
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDuration)
	prometheus.MustRegister(HTTPActiveRequests)

	// DB
	prometheus.MustRegister(DBQueryDuration)
	prometheus.MustRegister(DBQueryTotal)
	prometheus.MustRegister(DBSlowQueries)
	prometheus.MustRegister(DBPoolSize)

	// Kafka
	prometheus.MustRegister(KafkaMessagesTotal)
	prometheus.MustRegister(KafkaPublishDuration)

	// Business
	prometheus.MustRegister(AccountsCreatedTotal)
	prometheus.MustRegister(AccountsClosedTotal)
	prometheus.MustRegister(TransfersTotal)
	prometheus.MustRegister(DepositsTotal)
	prometheus.MustRegister(WithdrawalsTotal)
	prometheus.MustRegister(LoginsTotal)
}
