package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	StatusHealthy   = "healthy"
	StatusDegraded  = "degraded"
	StatusUnhealthy = "unhealthy"
)

// Dependency represents a service dependency that can be health checked
type Dependency interface {
	Name() string
	HealthCheck(ctx context.Context) error
}

// ComponentStatus represents health status of a single component
type ComponentStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// HealthStatus represents aggregated health status
type HealthStatus struct {
	Status     string            `json:"status"`
	Components []ComponentStatus `json:"components"`
	Timestamp  time.Time         `json:"timestamp"`
}

// Checker aggregates health checks from multiple dependencies
type Checker struct {
	dependencies []Dependency
}

// NewChecker initializes health checker with dependency references
func NewChecker(deps ...Dependency) *Checker {
	return &Checker{
		dependencies: deps,
	}
}

// Check executes all health checks concurrently
func (c *Checker) Check(ctx context.Context) (HealthStatus, error) {
	if len(c.dependencies) == 0 {
		return HealthStatus{
			Status:     StatusHealthy,
			Components: []ComponentStatus{},
			Timestamp:  time.Now(),
		}, nil
	}

	type result struct {
		name string
		err  error
	}

	results := make(chan result, len(c.dependencies))
	var wg sync.WaitGroup

	// Execute health checks concurrently
	for _, dep := range c.dependencies {
		wg.Add(1)
		go func(d Dependency) {
			defer wg.Done()
			err := d.HealthCheck(ctx)
			results <- result{name: d.Name(), err: err}
		}(dep)
	}

	wg.Wait()
	close(results)

	// Aggregate results
	components := make([]ComponentStatus, 0, len(c.dependencies))
	healthyCount := 0
	unhealthyCount := 0

	for r := range results {
		comp := ComponentStatus{
			Name: r.name,
		}

		if r.err != nil {
			comp.Status = StatusUnhealthy
			comp.Error = r.err.Error()
			unhealthyCount++
		} else {
			comp.Status = StatusHealthy
			healthyCount++
		}

		components = append(components, comp)
	}

	// Determine overall status
	overallStatus := StatusHealthy
	var err error

	if unhealthyCount > 0 {
		if healthyCount > 0 {
			overallStatus = StatusDegraded
		} else {
			overallStatus = StatusUnhealthy
		}
		err = fmt.Errorf("%d of %d components unhealthy", unhealthyCount, len(c.dependencies))
	}

	return HealthStatus{
		Status:     overallStatus,
		Components: components,
		Timestamp:  time.Now(),
	}, err
}

// HTTPHandler returns HTTP handler for /healthz endpoint
func (c *Checker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		status, err := c.Check(ctx)

		w.Header().Set("Content-Type", "application/json")

		// Return 503 if unhealthy or degraded
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err := json.NewEncoder(w).Encode(status); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
		}
	}
}
