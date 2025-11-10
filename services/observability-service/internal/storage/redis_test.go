package storage

import (
	"context"
	"testing"
	"time"
)

func TestNewRedisClient_ConnectionFailed(t *testing.T) {
	// Unreachable Redis returns connection refused
	_, err := NewRedisClient("localhost:9999", "", 0)
	if err == nil {
		t.Fatal("expected connection error for unreachable Redis")
	}
}

func TestRedisClient_Close(t *testing.T) {
	// Test that Close doesn't panic on nil client
	client := &RedisClient{}
	err := client.Close()
	if err != nil {
		t.Errorf("Close on nil client should not error, got: %v", err)
	}
}

func TestRedisClient_HealthCheck_NilClient(t *testing.T) {
	// Nil client returns error
	client := &RedisClient{}
	ctx := context.Background()

	err := client.HealthCheck(ctx)
	if err == nil {
		t.Error("expected error for nil client")
	}
}

func TestRedisClient_HealthCheck_Timeout(t *testing.T) {
	// Integration test - skipped
	t.Skip("integration test: requires Redis")

	client, err := NewRedisClient("localhost:6379", "", 0)
	if err != nil {
		t.Skip("cannot connect to test Redis")
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	err = client.HealthCheck(ctx)
	if err == nil {
		t.Error("expected timeout error")
	}
}
