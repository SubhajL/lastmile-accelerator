package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock dependency for testing
type mockDependency struct {
	name    string
	healthy bool
	err     error
}

func (m *mockDependency) Name() string {
	return m.name
}

func (m *mockDependency) HealthCheck(ctx context.Context) error {
	if !m.healthy {
		return m.err
	}
	return nil
}

func TestCheck_AllHealthy(t *testing.T) {
	// All dependencies up returns healthy status
	deps := []Dependency{
		&mockDependency{name: "database", healthy: true},
		&mockDependency{name: "redis", healthy: true},
	}

	checker := NewChecker(deps...)
	status, err := checker.Check(context.Background())

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if status.Status != StatusHealthy {
		t.Errorf("expected status %s, got %s", StatusHealthy, status.Status)
	}

	if len(status.Components) != 2 {
		t.Errorf("expected 2 components, got %d", len(status.Components))
	}
}

func TestCheck_OneUnhealthy(t *testing.T) {
	// DB down returns degraded with details
	deps := []Dependency{
		&mockDependency{name: "database", healthy: false, err: fmt.Errorf("connection refused")},
		&mockDependency{name: "redis", healthy: true},
	}

	checker := NewChecker(deps...)
	status, err := checker.Check(context.Background())

	if err == nil {
		t.Fatal("expected error when dependency unhealthy")
	}

	if status.Status != StatusDegraded {
		t.Errorf("expected status %s, got %s", StatusDegraded, status.Status)
	}

	// Check database component is unhealthy
	var dbComponent *ComponentStatus
	for i := range status.Components {
		if status.Components[i].Name == "database" {
			dbComponent = &status.Components[i]
			break
		}
	}

	if dbComponent == nil {
		t.Fatal("database component not found")
	}

	if dbComponent.Status != StatusUnhealthy {
		t.Errorf("expected database status %s, got %s", StatusUnhealthy, dbComponent.Status)
	}
}

func TestCheck_AllUnhealthy(t *testing.T) {
	// All deps down returns unhealthy status
	deps := []Dependency{
		&mockDependency{name: "database", healthy: false, err: fmt.Errorf("db error")},
		&mockDependency{name: "redis", healthy: false, err: fmt.Errorf("redis error")},
	}

	checker := NewChecker(deps...)
	status, err := checker.Check(context.Background())

	if err == nil {
		t.Fatal("expected error when all dependencies unhealthy")
	}

	if status.Status != StatusUnhealthy {
		t.Errorf("expected status %s, got %s", StatusUnhealthy, status.Status)
	}
}

func TestHTTPHandler_ReturnsJSON(t *testing.T) {
	// Handler responds 200 with JSON body
	deps := []Dependency{
		&mockDependency{name: "database", healthy: true},
	}

	checker := NewChecker(deps...)
	handler := checker.HTTPHandler()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var response HealthStatus
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Status != StatusHealthy {
		t.Errorf("expected healthy status in response, got %s", response.Status)
	}
}

func TestHTTPHandler_ServiceUnavailable(t *testing.T) {
	// Unhealthy dependencies return 503 status
	deps := []Dependency{
		&mockDependency{name: "database", healthy: false, err: fmt.Errorf("db down")},
	}

	checker := NewChecker(deps...)
	handler := checker.HTTPHandler()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", rec.Code)
	}
}
