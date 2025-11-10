package cache

import (
	"context"
	"testing"
)

func TestNewRedisClient_InvalidAddr_ReturnsError(t *testing.T) {
	// Arrange
	invalidAddr := ""

	// Act
	_, err := NewRedisClient(invalidAddr)

	// Assert
	if err == nil {
		t.Fatal("expected error for empty address, got nil")
	}
}

func TestHealthCheck_NilClient_ReturnsError(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	err := HealthCheck(ctx, nil)

	// Assert
	if err == nil {
		t.Fatal("expected error for nil client, got nil")
	}
}

func TestHealthCheck_UnhealthyRedis_ReturnsError(t *testing.T) {
	// Arrange - create client with unreachable address
	ctx := context.Background()
	client, _ := NewRedisClient("localhost:9999")

	// Act
	err := HealthCheck(ctx, client)

	// Assert
	if err == nil {
		t.Fatal("expected error for unreachable Redis, got nil")
	}
}
