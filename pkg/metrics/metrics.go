package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var WorkflowExecutionsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "tempo_workflow_executions_total",
		Help: "Total number of workflow executions",
	},
	[]string{"workflow_id", "status"}, // Labels để filter
)

var WorkflowExecutionDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "tempo_workflow_execution_duration_seconds",
		Help: "Workflow execution duration in seconds",
		// Buckets định nghĩa các khoảng thời gian
		// Default: [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
		Buckets: prometheus.DefBuckets,
	},
	[]string{"workflow_id"},
)
var WorkflowExecutionsActive = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "tempo_workflow_executions_active",
		Help: "Number of currently running workflow executions",
	},
)

// ActionExecutionsTotal đếm số lần actions được execute
var ActionExecutionsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "tempo_action_executions_total",
		Help: "Total number of action executions",
	},
	[]string{"action_type", "status"}, // http, email, etc
)

var ActionExecutionDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "tempo_action_execution_duration_seconds",
		Help:    "Action execution duration in seconds",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	},
	[]string{"action_type"},
)
var HTTPRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "tempo_http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "endpoint", "status"},
)

var HTTPRequestDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "tempo_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	},
	[]string{"method", "endpoint"},
)

var HTTPRequestsInFlight = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "tempo_http_requests_in_flight",
		Help: "Number of HTTP requests currently being processed",
	},
)

// DBQueriesTotal đếm số database queries
var DBQueriesTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "tempo_db_queries_total",
		Help: "Total number of database queries",
	},
	[]string{"operation", "table"},
)

var DBQueryDuration = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "tempo_db_query_duration_seconds",
		Help:    "Database query duration in seconds",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
	},
	[]string{"operation", "table"},
)

var ConnectorExecutionsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "tempo_connector_executions_total",
		Help: "Total number of connector executions",
	},
	[]string{"connector_type", "status"},
)

var ConnectorCircuitBreakerState = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "tempo_connector_circuit_breaker_state",
		Help: "Circuit breaker state (0=closed, 1=open, 2=half-open)",
	},
	[]string{"connector_type"},
)

// GoRoutinesCount đếm số goroutines
var GoRoutinesCount = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "tempo_goroutines_count",
		Help: "Number of goroutines",
	},
)

// MemoryUsageBytes đo memory usage
var MemoryUsageBytes = promauto.NewGauge(
	prometheus.GaugeOpts{
		Name: "tempo_memory_usage_bytes",
		Help: "Memory usage in bytes",
	},
)
