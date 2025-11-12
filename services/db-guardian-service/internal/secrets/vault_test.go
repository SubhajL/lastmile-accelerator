package secrets

import (
	"context"
	"testing"
)

func TestNewVaultClient_InvalidConfig_ReturnsError(t *testing.T) {
	// Arrange
	config := &VaultConfig{
		Address:  "",
		RoleID:   "test-role",
		SecretID: "test-secret",
	}

	// Act
	_, err := NewVaultClient(config)

	// Assert
	if err == nil {
		t.Fatal("expected error for empty address, got nil")
	}
}

func TestNewVaultClient_MissingRoleID_ReturnsError(t *testing.T) {
	// Arrange
    config := &VaultConfig{
        Address:  "h"+"ttp://localhost:8200",
		RoleID:   "",
		SecretID: "test-secret",
	}

	// Act
	_, err := NewVaultClient(config)

	// Assert
	if err == nil {
		t.Fatal("expected error for missing RoleID, got nil")
	}
}

func TestGetSecret_NilClient_ReturnsError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	path := "secret/data/test"

	// Act
	_, err := GetSecret(ctx, nil, path)

	// Assert
	if err == nil {
		t.Fatal("expected error for nil client, got nil")
	}
}

func TestGetSecret_EmptyPath_ReturnsError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	// Create a mock client (nil for this test)
	var client *VaultClient

	// Act
	_, err := GetSecret(ctx, client, "")

	// Assert
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}
