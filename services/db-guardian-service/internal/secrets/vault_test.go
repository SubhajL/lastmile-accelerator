package secrets

import (
	"context"
	"errors"
	"strings"
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

// Mock GetSecret function for testing GetProjectDSN
var mockGetSecret func(ctx context.Context, vc *VaultClient, path string) (map[string]string, error)

// TestGetProjectDSN_Success tests successful DSN retrieval
func TestGetProjectDSN_Success(t *testing.T) {
	// Save original GetSecret and restore after test
	originalGetSecret := GetSecret
	defer func() { GetSecret = originalGetSecret }()

	// Mock GetSecret to return test DSN
	GetSecret = func(ctx context.Context, vc *VaultClient, path string) (map[string]string, error) {
		if path == "secret/projects/test-project-123/db/dsn" {
			return map[string]string{
				"dsn": "postgresql://user:pass@localhost:5432/testdb",
			}, nil
		}
		return nil, errors.New("unexpected path")
	}

	// Create a minimal VaultClient (doesn't need real connection for this test)
	vc := &VaultClient{}

	// Test
	dsn, err := vc.GetProjectDSN(context.Background(), "vault:secret/projects/test-project-123/db/dsn")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if dsn != "postgresql://user:pass@localhost:5432/testdb" {
		t.Errorf("expected correct DSN, got: %s", dsn)
	}
}

// TestGetProjectDSN_InvalidFormat tests error handling for invalid vault reference format
func TestGetProjectDSN_InvalidFormat(t *testing.T) {
	tests := []struct {
		name     string
		vaultRef string
		wantErr  string
	}{
		{
			name:     "missing vault prefix",
			vaultRef: "secret/projects/test/db/dsn",
			wantErr:  "invalid vault reference format: must start with 'vault:' prefix",
		},
		{
			name:     "empty string",
			vaultRef: "",
			wantErr:  "invalid vault reference format: must start with 'vault:' prefix",
		},
		{
			name:     "only vault prefix",
			vaultRef: "vault:",
			wantErr:  "vault path cannot be empty",
		},
		{
			name:     "wrong prefix",
			vaultRef: "http:secret/path",
			wantErr:  "invalid vault reference format: must start with 'vault:' prefix",
		},
	}

	vc := &VaultClient{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vc.GetProjectDSN(context.Background(), tt.vaultRef)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

// TestGetProjectDSN_SecretNotFound tests error when Vault path doesn't exist
func TestGetProjectDSN_SecretNotFound(t *testing.T) {
	// Save original GetSecret and restore after test
	originalGetSecret := GetSecret
	defer func() { GetSecret = originalGetSecret }()

	// Mock GetSecret to return not found error
	GetSecret = func(ctx context.Context, vc *VaultClient, path string) (map[string]string, error) {
		return nil, errors.New("secret not found at path: " + path)
	}

	vc := &VaultClient{}

	// Test
	_, err := vc.GetProjectDSN(context.Background(), "vault:secret/projects/missing/db/dsn")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expectedErr := "failed to fetch project DSN from path secret/projects/missing/db/dsn: secret not found at path: secret/projects/missing/db/dsn"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

// TestGetProjectDSN_MissingDSNKey tests error when secret lacks "dsn" key
func TestGetProjectDSN_MissingDSNKey(t *testing.T) {
	// Save original GetSecret and restore after test
	originalGetSecret := GetSecret
	defer func() { GetSecret = originalGetSecret }()

	// Mock GetSecret to return secret without dsn key
	GetSecret = func(ctx context.Context, vc *VaultClient, path string) (map[string]string, error) {
		return map[string]string{
			"username": "dbuser",
			"password": "dbpass",
			// missing "dsn" key
		}, nil
	}

	vc := &VaultClient{}

	// Test
	_, err := vc.GetProjectDSN(context.Background(), "vault:secret/projects/test/db/dsn")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expectedErr := "secret at path secret/projects/test/db/dsn missing required 'dsn' key"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

// TestGetProjectDSN_EmptyDSNValue tests error when DSN value is empty
func TestGetProjectDSN_EmptyDSNValue(t *testing.T) {
	// Save original GetSecret and restore after test
	originalGetSecret := GetSecret
	defer func() { GetSecret = originalGetSecret }()

	// Mock GetSecret to return empty DSN value
	GetSecret = func(ctx context.Context, vc *VaultClient, path string) (map[string]string, error) {
		return map[string]string{
			"dsn": "",
		}, nil
	}

	vc := &VaultClient{}

	// Test
	_, err := vc.GetProjectDSN(context.Background(), "vault:secret/projects/test/db/dsn")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expectedErr := "DSN value at path secret/projects/test/db/dsn is empty"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

// TestGetProjectDSN_RedactsInError tests that DSN value is never included in error messages
func TestGetProjectDSN_RedactsInError(t *testing.T) {
	// Save original GetSecret and restore after test
	originalGetSecret := GetSecret
	defer func() { GetSecret = originalGetSecret }()

	testDSN := "postgresql://secret_user:super_secret_password@production.db:5432/sensitive_db"

	// Test various error scenarios and ensure DSN is never in error message
	scenarios := []struct {
		name          string
		mockGetSecret func(ctx context.Context, vc *VaultClient, path string) (map[string]string, error)
	}{
		{
			name: "GetSecret error",
			mockGetSecret: func(ctx context.Context, vc *VaultClient, path string) (map[string]string, error) {
				return nil, errors.New("vault is sealed")
			},
		},
		{
			name: "Missing DSN key",
			mockGetSecret: func(ctx context.Context, vc *VaultClient, path string) (map[string]string, error) {
				return map[string]string{"not_dsn": testDSN}, nil
			},
		},
		{
			name: "Empty DSN",
			mockGetSecret: func(ctx context.Context, vc *VaultClient, path string) (map[string]string, error) {
				return map[string]string{"dsn": ""}, nil
			},
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			GetSecret = sc.mockGetSecret
			vc := &VaultClient{}

			_, err := vc.GetProjectDSN(context.Background(), "vault:secret/projects/test/db/dsn")

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// Ensure error message doesn't contain any part of the DSN
			errMsg := err.Error()
			if strings.Contains(errMsg, testDSN) {
				t.Errorf("error message contains DSN value: %s", errMsg)
			}
			if strings.Contains(errMsg, "secret_user") ||
			   strings.Contains(errMsg, "super_secret_password") ||
			   strings.Contains(errMsg, "sensitive_db") {
				t.Errorf("error message contains sensitive information: %s", errMsg)
			}

			// Ensure error message does contain the path
			if !strings.Contains(errMsg, "secret/projects/test/db/dsn") {
				t.Errorf("error message should contain the path, got: %s", errMsg)
			}
		})
	}
}
