package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/lma/db-guardian-service/internal/config"
)

func TestNew_RegistersHealthEndpoint(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		ServiceName: "test-service",
		ServicePort: "7105",
		Environment: "test",
	}
	deps := &Dependencies{}

	// Act
	srv := New(cfg, deps)

	// Assert
	if srv == nil {
		t.Fatal("expected server instance, got nil")
	}

	// Test that healthz route exists
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	srv.handler.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Error("expected /healthz route to exist")
	}
}

func TestStart_ListensOnConfiguredPort(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		ServiceName: "test-service",
		ServicePort: "0", // Use port 0 to get random available port
		Environment: "test",
	}
	deps := &Dependencies{}
	srv := New(cfg, deps)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Act
	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(10 * time.Millisecond)

	// Assert - cancel context to trigger shutdown
	cancel()

	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down gracefully")
	}
}

func TestStart_ContextCanceled_ShutsDownGracefully(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		ServiceName: "test-service",
		ServicePort: "0",
		Environment: "test",
	}
	deps := &Dependencies{}
	srv := New(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	// Act
	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Start(ctx)
	}()

	time.Sleep(10 * time.Millisecond)
	cancel()

	// Assert
	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("expected clean shutdown, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down within timeout")
	}
}
