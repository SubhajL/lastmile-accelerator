package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestHealthHandler_AllHealthy_Returns200(t *testing.T) {
	// Arrange - mock healthy dependencies
	deps := &HealthDeps{
		DBPing: func(ctx context.Context, db *sql.DB) error {
			return nil
		},
		RedisPing: func(ctx context.Context, client *redis.Client) error {
			return nil
		},
	}

	handler := NewHealthHandler(deps)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	// Act
	handler(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%v'", response["status"])
	}
}

func TestHealthHandler_DBUnhealthy_Returns503(t *testing.T) {
	// Arrange - mock unhealthy DB
	deps := &HealthDeps{
		DBPing: func(ctx context.Context, db *sql.DB) error {
			return fmt.Errorf("database connection failed")
		},
		RedisPing: func(ctx context.Context, client *redis.Client) error {
			return nil
		},
	}

	handler := NewHealthHandler(deps)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	// Act
	handler(w, req)

	// Assert
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["status"] != "unhealthy" {
		t.Errorf("expected status 'unhealthy', got '%v'", response["status"])
	}
}

func TestHealthHandler_PartialFailure_IncludesDetails(t *testing.T) {
	// Arrange - mixed health status
	deps := &HealthDeps{
		DBPing: func(ctx context.Context, db *sql.DB) error {
			return nil
		},
		RedisPing: func(ctx context.Context, client *redis.Client) error {
			return fmt.Errorf("redis connection failed")
		},
	}

	handler := NewHealthHandler(deps)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	// Act
	handler(w, req)

	// Assert
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	checks, ok := response["checks"].(map[string]interface{})
	if !ok {
		t.Fatal("expected checks in response")
	}

	if dbStatus, ok := checks["database"].(string); !ok || dbStatus != "ok" {
		t.Errorf("expected database status 'ok', got '%v'", checks["database"])
	}

	if redisStatus, ok := checks["redis"].(string); !ok || redisStatus == "ok" {
		t.Errorf("expected redis status to be unhealthy, got '%v'", checks["redis"])
	}
}
