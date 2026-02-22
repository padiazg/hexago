package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusUnknown   HealthStatus = "unknown"
)

// ComponentHealth represents the health state of an individual component
type ComponentHealth struct {
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	Timestamp time.Time    `json:"timestamp"`
}

// HealthReport is the complete system health report
type HealthReport struct {
	Status     HealthStatus               `json:"status"`
	Components map[string]ComponentHealth `json:"components"`
	Timestamp  time.Time                  `json:"timestamp"`
	Uptime     string                     `json:"uptime"`
}

// HealthChecker manages system health checks
type HealthChecker struct {
	mu         sync.RWMutex
	components map[string]ComponentHealth
	startTime  time.Time
	checks     map[string]func(context.Context) ComponentHealth
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		components: make(map[string]ComponentHealth),
		startTime:  time.Now(),
		checks:     make(map[string]func(context.Context) ComponentHealth),
	}
}

// RegisterCheck registers a health check function for a component
func (h *HealthChecker) RegisterCheck(name string, check func(context.Context) ComponentHealth) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checks[name] = check
}

// UpdateComponentHealth updates the health status of a component
func (h *HealthChecker) UpdateComponentHealth(name string, health ComponentHealth) {
	h.mu.Lock()
	defer h.mu.Unlock()
	health.Timestamp = time.Now()
	h.components[name] = health
}

// GetHealthReport generates a complete health report by executing all checks
func (h *HealthChecker) GetHealthReport(ctx context.Context) HealthReport {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Execute all registered checks
	for name, check := range h.checks {
		h.components[name] = check(ctx)
	}

	// Determine overall status
	overallStatus := StatusHealthy
	for _, comp := range h.components {
		if comp.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		}
		if comp.Status == StatusUnknown && overallStatus == StatusHealthy {
			overallStatus = StatusUnknown
		}
	}

	// Copy components to avoid exposing internal map
	componentsCopy := make(map[string]ComponentHealth, len(h.components))
	for k, v := range h.components {
		componentsCopy[k] = v
	}

	return HealthReport{
		Status:     overallStatus,
		Components: componentsCopy,
		Timestamp:  time.Now(),
		Uptime:     time.Since(h.startTime).String(),
	}
}

// HTTPHandler returns an http.Handler for the health endpoint
func (h *HealthChecker) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		report := h.GetHealthReport(ctx)

		w.Header().Set("Content-Type", "application/json")

		// HTTP status code based on health
		if report.Status == StatusHealthy {
			w.WriteHeader(http.StatusOK)
		} else if report.Status == StatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK) // Unknown still responds 200
		}

		json.NewEncoder(w).Encode(report)
	})
}

// ReadinessHandler returns a simple handler for readiness (ready to receive traffic)
func (h *HealthChecker) ReadinessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		report := h.GetHealthReport(ctx)

		if report.Status == StatusHealthy {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "ready")
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, "not ready")
		}
	})
}

// LivenessHandler returns a simple handler for liveness (process alive)
func LivenessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "alive")
	})
}
