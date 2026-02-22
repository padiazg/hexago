package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics contains all Prometheus metrics
type PrometheusMetrics struct {
	// Request metrics
	RequestsTotal   prometheus.Counter
	RequestsSuccess prometheus.Counter
	RequestsFailed  prometheus.Counter

	// Latency histogram
	RequestLatency prometheus.Histogram

	// Active operations gauge
	ActiveOperations prometheus.Gauge
}

// NewPrometheusMetrics creates and initializes Prometheus metrics
func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		RequestsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "app_requests_total",
			Help: "Total number of requests processed",
		}),
		RequestsSuccess: promauto.NewCounter(prometheus.CounterOpts{
			Name: "app_requests_success_total",
			Help: "Total number of successful requests",
		}),
		RequestsFailed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "app_requests_failed_total",
			Help: "Total number of failed requests",
		}),
		RequestLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "app_request_latency_seconds",
			Help:    "Request latency in seconds",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 12), // 1ms to ~4s
		}),
		ActiveOperations: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "app_active_operations",
			Help: "Number of currently active operations",
		}),
	}
}

// RecordRequest is a helper to record a request with success/failure status
func (m *PrometheusMetrics) RecordRequest(operation string, status string) {
	if m == nil {
		return
	}

	m.RequestsTotal.Inc()

	if status == "success" {
		m.RequestsSuccess.Inc()
	} else {
		m.RequestsFailed.Inc()
	}
}
