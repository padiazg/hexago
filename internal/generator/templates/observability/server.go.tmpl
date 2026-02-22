package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server manages observability endpoints (health, metrics)
type Server struct {
	server  *http.Server
	health  *HealthChecker
	metrics *PrometheusMetrics
}

// NewServer creates a new observability server
func NewServer(addr string, health *HealthChecker, metrics *PrometheusMetrics) *Server {
	mux := http.NewServeMux()

	s := &Server{
		health:  health,
		metrics: metrics,
		server: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}

	// Register health endpoints
	mux.Handle("/health", health.HTTPHandler())
	mux.Handle("/health/ready", health.ReadinessHandler())
	mux.Handle("/health/live", LivenessHandler())

	// Register metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	return s
}

// Start starts the observability server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// StartInBackground starts the server in a background goroutine and returns a shutdown function
func (s *Server) StartInBackground(logger interface{ Info(string, ...interface{}) }) func() {
	go func() {
		logger.Info("Observability server started on %s", s.server.Addr)
		logger.Info("  Health:            http://%s/health", s.server.Addr)
		logger.Info("  Ready:             http://%s/health/ready", s.server.Addr)
		logger.Info("  Live:              http://%s/health/live", s.server.Addr)
		logger.Info("  Metrics (Prom):    http://%s/metrics", s.server.Addr)

		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			logger.Info("Observability server error: %v", err)
		}
	}()

	// Return shutdown function
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			fmt.Printf("Error shutting down observability server: %v\n", err)
		}
	}
}
