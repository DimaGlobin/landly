package metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const namespace = "landly"

// HTTP metrics - collected automatically by middleware
var (
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latency in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	HTTPRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_requests_in_flight",
			Help:      "Current number of HTTP requests being processed",
		},
	)
)

// Business metrics
var (
	UserRegistrationsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "user_registrations_total",
			Help:      "Total number of user registrations",
		},
	)

	UserLoginsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "user_logins_total",
			Help:      "Total number of user logins",
		},
	)

	UserLoginFailuresTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "user_login_failures_total",
			Help:      "Total number of failed login attempts",
		},
	)

	ProjectsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "projects_created_total",
			Help:      "Total number of projects created",
		},
	)

	ProjectsDeletedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "projects_deleted_total",
			Help:      "Total number of projects deleted",
		},
	)

	GenerationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "generations_total",
			Help:      "Total number of site generations",
		},
		[]string{"type", "status"},
	)

	GenerationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "generation_duration_seconds",
			Help:      "Site generation duration in seconds",
			Buckets:   []float64{0.5, 1, 2, 5, 10, 30, 60, 120},
		},
		[]string{"type"},
	)

	PublishOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "publish_operations_total",
			Help:      "Total number of publish/unpublish operations",
		},
		[]string{"action"},
	)
)

// Database metrics
var (
	DBPoolOpenConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_pool_open_connections",
			Help:      "Number of open database connections",
		},
	)

	DBPoolInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_pool_in_use",
			Help:      "Number of database connections currently in use",
		},
	)

	DBPoolIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_pool_idle",
			Help:      "Number of idle database connections",
		},
	)

	DBPoolWaitCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "db_pool_wait_count_total",
			Help:      "Total number of connections waited for",
		},
	)
)

// RecordDBStats updates database pool metrics from sql.DBStats
func RecordDBStats(stats sql.DBStats) {
	DBPoolOpenConnections.Set(float64(stats.OpenConnections))
	DBPoolInUse.Set(float64(stats.InUse))
	DBPoolIdle.Set(float64(stats.Idle))
	DBPoolWaitCount.Set(float64(stats.WaitCount))
}

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, path, status string, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordGeneration records generation metrics
func RecordGeneration(genType, status string, duration float64) {
	GenerationsTotal.WithLabelValues(genType, status).Inc()
	GenerationDuration.WithLabelValues(genType).Observe(duration)
}

// RecordPublish records publish operation
func RecordPublish(action string) {
	PublishOperationsTotal.WithLabelValues(action).Inc()
}

