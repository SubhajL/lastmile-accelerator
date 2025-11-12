package vault

import (
	"context"
	"testing"

	"example.com/lma/secrets-env-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient_ValidConfig(t *testing.T) {
	cfg := &config.VaultConfig{
    Address:   "h"+"ttp://localhost:8200",
		RoleID:    "test-role",
		SecretID:  "test-secret",
		MountPath: "secret",
	}

	client, err := NewClient(cfg)
	
	// In test mode (no real Vault), should still create client
	require.NoError(t, err)
	assert.NotNil(t, client)
}

func TestWriteSecret_Success(t *testing.T) {
	client := &Client{
		testMode: true,
		testData: make(map[string]map[string]interface{}),
	}

	ctx := context.Background()
	data := map[string]interface{}{
		"username": "admin",
		"password": "secret123",
	}

	err := client.WriteSecret(ctx, "test/path", data)
	assert.NoError(t, err)
}

func TestReadSecret_ExistingSecret(t *testing.T) {
	client := &Client{
		testMode: true,
		testData: map[string]map[string]interface{}{
			"test/path": {
				"username": "admin",
				"password": "secret123",
			},
		},
	}

	ctx := context.Background()
	data, err := client.ReadSecret(ctx, "test/path")
	
	require.NoError(t, err)
	assert.Equal(t, "admin", data["username"])
	assert.Equal(t, "secret123", data["password"])
}

func TestReadSecret_NotFound(t *testing.T) {
	client := &Client{
		testMode: true,
		testData: make(map[string]map[string]interface{}),
	}

	ctx := context.Background()
	_, err := client.ReadSecret(ctx, "missing/path")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteSecret_ExistingSecret(t *testing.T) {
	client := &Client{
		testMode: true,
		testData: map[string]map[string]interface{}{
			"test/path": {"key": "value"},
		},
	}

	ctx := context.Background()
	err := client.DeleteSecret(ctx, "test/path")
	
	assert.NoError(t, err)
	
	// Verify it's deleted
	_, err = client.ReadSecret(ctx, "test/path")
	assert.Error(t, err)
}

func TestDeleteSecret_Idempotent(t *testing.T) {
	client := &Client{
		testMode: true,
		testData: make(map[string]map[string]interface{}),
	}

	ctx := context.Background()
	
	// Delete non-existent secret should not error
	err := client.DeleteSecret(ctx, "missing/path")
	assert.NoError(t, err)
	
	// Delete again should also not error
	err = client.DeleteSecret(ctx, "missing/path")
	assert.NoError(t, err)
}

func TestListSecrets_WithResults(t *testing.T) {
	client := &Client{
		testMode: true,
		testData: map[string]map[string]interface{}{
			"prefix/secret1": {"key": "value1"},
			"prefix/secret2": {"key": "value2"},
			"other/secret3":  {"key": "value3"},
		},
	}

	ctx := context.Background()
	secrets, err := client.ListSecrets(ctx, "prefix/")
	
	require.NoError(t, err)
	assert.Len(t, secrets, 2)
	assert.Contains(t, secrets, "prefix/secret1")
	assert.Contains(t, secrets, "prefix/secret2")
}

func TestListSecrets_EmptyPath(t *testing.T) {
	client := &Client{
		testMode: true,
		testData: make(map[string]map[string]interface{}),
	}

	ctx := context.Background()
	secrets, err := client.ListSecrets(ctx, "prefix/")
	
	require.NoError(t, err)
	assert.Empty(t, secrets)
}

func TestHealthCheck_VaultAvailable(t *testing.T) {
	client := &Client{
		testMode: true,
	}

	ctx := context.Background()
	err := client.HealthCheck(ctx)
	
	assert.NoError(t, err)
}

func TestClose_NoError(t *testing.T) {
	client := &Client{
		testMode: true,
	}

	err := client.Close()
	assert.NoError(t, err)
}
