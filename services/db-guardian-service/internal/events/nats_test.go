package events

import (
	"context"
	"testing"
)

func TestNewNATSClient_InvalidURL_ReturnsError(t *testing.T) {
	// Arrange
	invalidURL := ""

	// Act
	_, err := NewNATSClient(invalidURL)

	// Assert
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

func TestNewNATSClient_UnreachableServer_ReturnsError(t *testing.T) {
	// Arrange
	unreachableURL := "nats://localhost:9999"

	// Act
	_, err := NewNATSClient(unreachableURL)

	// Assert
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
}

func TestPublishEvent_NilConnection_ReturnsError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	subject := "test.subject"
	data := []byte("test data")

	// Act
	err := PublishEvent(ctx, nil, subject, data)

	// Assert
	if err == nil {
		t.Fatal("expected error for nil connection, got nil")
	}
}
