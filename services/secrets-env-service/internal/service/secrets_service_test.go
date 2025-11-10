package service

import (
	"context"
	"testing"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"example.com/lma/secrets-env-service/internal/repository"
	"example.com/lma/secrets-env-service/internal/vault"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSecret_Success(t *testing.T) {
	vaultClient := &vault.Client{}
	vaultClient.SetTestMode(true)
	
	repo := repository.NewSecretsRepository(nil)
	
	service := NewSecretsService(vaultClient, repo, nil, nil)

	secret := &domain.Secret{
		ID:          uuid.New().String(),
		TenantID:    "tenant-1",
		ProjectID:   "proj-1",
		Key:         "API_KEY",
		Environment: "production",
		CreatedBy:   "user@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	value := map[string]interface{}{
		"api_key": "secret-value-123",
	}

	err := service.CreateSecret(context.Background(), secret, value)
	assert.NoError(t, err)
}

func TestCreateSecret_VaultFailure(t *testing.T) {
	vaultClient := &vault.Client{}
	vaultClient.SetTestMode(true)
	vaultClient.SetFailMode(true) // Simulate failure
	
	repo := repository.NewSecretsRepository(nil)
service := NewSecretsService(vaultClient, repo, nil, nil)

	secret := &domain.Secret{
		ID:          uuid.New().String(),
		TenantID:    "tenant-1",
		ProjectID:   "proj-1",
		Key:         "API_KEY",
		Environment: "production",
		CreatedBy:   "user@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := service.CreateSecret(context.Background(), secret, map[string]interface{}{})
	assert.Error(t, err)
}

func TestGetSecret_ExistingSecret(t *testing.T) {
	vaultClient := &vault.Client{}
	vaultClient.SetTestMode(true)
	
	repo := repository.NewSecretsRepository(nil)
service := NewSecretsService(vaultClient, repo, nil, nil)

	// Create first
	secret := &domain.Secret{
		ID:          uuid.New().String(),
		TenantID:    "tenant-1",
		ProjectID:   "proj-1",
		Key:         "DATABASE_URL",
		Environment: "staging",
		CreatedBy:   "user@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	value := map[string]interface{}{
		"connection_string": "postgres://localhost:5432/db",
	}

	err := service.CreateSecret(context.Background(), secret, value)
	require.NoError(t, err)

	// Get it back
	retrieved, retrievedValue, err := service.GetSecret(context.Background(), "tenant-1", "proj-1", "DATABASE_URL", "staging")
	require.NoError(t, err)
	assert.Equal(t, "DATABASE_URL", retrieved.Key)
	assert.Equal(t, "postgres://localhost:5432/db", retrievedValue["connection_string"])
}

func TestGetSecret_NotFound(t *testing.T) {
	vaultClient := &vault.Client{}
	vaultClient.SetTestMode(true)
	
	repo := repository.NewSecretsRepository(nil)
service := NewSecretsService(vaultClient, repo, nil, nil)

	_, _, err := service.GetSecret(context.Background(), "tenant-1", "proj-1", "NONEXISTENT", "production")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteSecret_Success(t *testing.T) {
	vaultClient := &vault.Client{}
	vaultClient.SetTestMode(true)
	
	repo := repository.NewSecretsRepository(nil)
service := NewSecretsService(vaultClient, repo, nil, nil)

	// Create first
	secret := &domain.Secret{
		ID:          uuid.New().String(),
		TenantID:    "tenant-1",
		ProjectID:   "proj-1",
		Key:         "TEMP_KEY",
		Environment: "dev",
		CreatedBy:   "user@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := service.CreateSecret(context.Background(), secret, map[string]interface{}{"key": "value"})
	require.NoError(t, err)

	// Delete it
	err = service.DeleteSecret(context.Background(), "tenant-1", "proj-1", "TEMP_KEY", "dev")
	assert.NoError(t, err)

	// Verify it's gone
	_, _, err = service.GetSecret(context.Background(), "tenant-1", "proj-1", "TEMP_KEY", "dev")
	assert.Error(t, err)
}

func TestListSecrets_FiltersCorrectly(t *testing.T) {
	vaultClient := &vault.Client{}
	vaultClient.SetTestMode(true)
	
	repo := repository.NewSecretsRepository(nil)
service := NewSecretsService(vaultClient, repo, nil, nil)

	// Create multiple secrets
	for i := 0; i < 3; i++ {
		secret := &domain.Secret{
			ID:          uuid.New().String(),
			TenantID:    "tenant-1",
			ProjectID:   "proj-1",
			Key:         "KEY_" + string(rune('A'+i)),
			Environment: "production",
			CreatedBy:   "user@example.com",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		service.CreateSecret(context.Background(), secret, map[string]interface{}{})
	}

	secrets, _, err := service.ListSecrets(context.Background(), "proj-1", "production", 10, "")
	require.NoError(t, err)
	assert.Len(t, secrets, 3)
}
