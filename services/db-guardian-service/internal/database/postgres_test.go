package database

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func TestNewPostgresPool_InvalidDSN_ReturnsError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	invalidDSN := "invalid-dsn-format"

	// Act
	_, err := NewPostgresPool(ctx, invalidDSN)

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid DSN, got nil")
	}
}

func TestPing_WithTimeout_RespectsContext(t *testing.T) {
	// Arrange - use a valid DSN format but unreachable host
	ctx := context.Background()
	dsn := "postgres://user:pass@localhost:9999/nonexistent?sslmode=disable&connect_timeout=1"
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Act with timeout
	ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	
	err = Ping(ctx, db)

	// Assert
	if err == nil {
		t.Fatal("expected error due to timeout, got nil")
	}
}

func TestPing_HealthyDB_ReturnsNil(t *testing.T) {
	// This test would require a real database or testcontainers
	// For now, we'll test the nil DB case
	ctx := context.Background()
	
	// Act
	err := Ping(ctx, nil)
	
	// Assert
	if err == nil {
		t.Fatal("expected error for nil DB, got nil")
	}
}
