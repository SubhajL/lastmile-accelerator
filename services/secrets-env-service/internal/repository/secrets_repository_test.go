package repository

import (
	"context"
	"testing"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock database for testing
type mockDB struct {
	secrets map[string]*domain.Secret
}

func newMockDB() *mockDB {
	return &mockDB{
		secrets: make(map[string]*domain.Secret),
	}
}

func TestCreate_ValidSecret(t *testing.T) {
	repo := NewSecretsRepository(nil) // Using test mode
	repo.testMode = true
	repo.testData = make(map[string]*domain.Secret)

	secret := &domain.Secret{
		ID:          "secret-123",
		TenantID:    "tenant-1",
		ProjectID:   "proj-1",
		Key:         "DATABASE_URL",
		Environment: "production",
		CreatedBy:   "user@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := repo.Create(context.Background(), secret)
	assert.NoError(t, err)
}

func TestCreate_DuplicateKey(t *testing.T) {
	repo := NewSecretsRepository(nil)
	repo.testMode = true
	repo.testData = make(map[string]*domain.Secret)

	secret := &domain.Secret{
		ID:          "secret-123",
		TenantID:    "tenant-1",
		ProjectID:   "proj-1",
		Key:         "DATABASE_URL",
		Environment: "production",
		CreatedBy:   "user@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// First create succeeds
	err := repo.Create(context.Background(), secret)
	require.NoError(t, err)

	// Duplicate should fail
	duplicate := &domain.Secret{
		ID:          "secret-456",
		TenantID:    "tenant-1",
		ProjectID:   "proj-1",
		Key:         "DATABASE_URL",
		Environment: "production",
		CreatedBy:   "user@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.Create(context.Background(), duplicate)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestGetByID_ExistingSecret(t *testing.T) {
	repo := NewSecretsRepository(nil)
	repo.testMode = true
	repo.testData = map[string]*domain.Secret{
		"secret-123": {
			ID:          "secret-123",
			TenantID:    "tenant-1",
			ProjectID:   "proj-1",
			Key:         "API_KEY",
			Environment: "staging",
			CreatedBy:   "user@example.com",
		},
	}

	secret, err := repo.GetByID(context.Background(), "secret-123")
	require.NoError(t, err)
	assert.Equal(t, "API_KEY", secret.Key)
	assert.Equal(t, "staging", secret.Environment)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := NewSecretsRepository(nil)
	repo.testMode = true
	repo.testData = make(map[string]*domain.Secret)

	_, err := repo.GetByID(context.Background(), "nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetByKey_NaturalKey(t *testing.T) {
	repo := NewSecretsRepository(nil)
	repo.testMode = true
	repo.testData = map[string]*domain.Secret{
		"secret-123": {
			ID:          "secret-123",
			TenantID:    "tenant-1",
			ProjectID:   "proj-1",
			Key:         "API_KEY",
			Environment: "production",
		},
	}

	secret, err := repo.GetByKey(context.Background(), "proj-1", "API_KEY", "production")
	require.NoError(t, err)
	assert.Equal(t, "secret-123", secret.ID)
}

func TestUpdate_ModifiesFields(t *testing.T) {
	repo := NewSecretsRepository(nil)
	repo.testMode = true
	
	old := time.Now().Add(-1 * time.Hour)
	repo.testData = map[string]*domain.Secret{
		"secret-123": {
			ID:          "secret-123",
			TenantID:    "tenant-1",
			ProjectID:   "proj-1",
			Key:         "API_KEY",
			Environment: "production",
			CreatedAt:   old,
			UpdatedAt:   old,
		},
	}

	updated := &domain.Secret{
		ID:        "secret-123",
		UpdatedAt: time.Now(),
	}

	err := repo.Update(context.Background(), updated)
	assert.NoError(t, err)

	result, _ := repo.GetByID(context.Background(), "secret-123")
	assert.True(t, result.UpdatedAt.After(old))
}

func TestDelete_RemovesRecord(t *testing.T) {
	repo := NewSecretsRepository(nil)
	repo.testMode = true
	repo.testData = map[string]*domain.Secret{
		"secret-123": {ID: "secret-123"},
	}

	err := repo.Delete(context.Background(), "secret-123")
	assert.NoError(t, err)

	_, err = repo.GetByID(context.Background(), "secret-123")
	assert.Error(t, err)
}

func TestList_OrderedByCreatedAt(t *testing.T) {
	repo := NewSecretsRepository(nil)
	repo.testMode = true
	
	now := time.Now()
	repo.testData = map[string]*domain.Secret{
		"secret-1": {
			ID:          "secret-1",
			ProjectID:   "proj-1",
			Environment: "production",
			CreatedAt:   now.Add(-2 * time.Hour),
		},
		"secret-2": {
			ID:          "secret-2",
			ProjectID:   "proj-1",
			Environment: "production",
			CreatedAt:   now.Add(-1 * time.Hour),
		},
		"secret-3": {
			ID:          "secret-3",
			ProjectID:   "proj-1",
			Environment: "production",
			CreatedAt:   now,
		},
	}

	secrets, _, err := repo.List(context.Background(), "proj-1", "production", 10, "")
	require.NoError(t, err)
	require.Len(t, secrets, 3)

	// Should be ordered newest first
	assert.Equal(t, "secret-3", secrets[0].ID)
	assert.Equal(t, "secret-2", secrets[1].ID)
	assert.Equal(t, "secret-1", secrets[2].ID)
}

func TestList_Pagination(t *testing.T) {
	repo := NewSecretsRepository(nil)
	repo.testMode = true

	// Create 5 secrets
	for i := 0; i < 5; i++ {
		id := string(rune('a' + i))
		repo.testData[id] = &domain.Secret{
			ID:          id,
			ProjectID:   "proj-1",
			Environment: "production",
			CreatedAt:   time.Now().Add(time.Duration(-i) * time.Hour),
		}
	}

	// Get first page
	secrets, cursor, err := repo.List(context.Background(), "proj-1", "production", 2, "")
	require.NoError(t, err)
	assert.Len(t, secrets, 2)
	assert.NotEmpty(t, cursor)

	// Get second page
	secrets, cursor, err = repo.List(context.Background(), "proj-1", "production", 2, cursor)
	require.NoError(t, err)
	assert.Len(t, secrets, 2)
}

func TestList_FiltersByEnvironment(t *testing.T) {
	repo := NewSecretsRepository(nil)
	repo.testMode = true
	repo.testData = map[string]*domain.Secret{
		"secret-1": {
			ID:          "secret-1",
			ProjectID:   "proj-1",
			Environment: "production",
		},
		"secret-2": {
			ID:          "secret-2",
			ProjectID:   "proj-1",
			Environment: "staging",
		},
	}

	secrets, _, err := repo.List(context.Background(), "proj-1", "production", 10, "")
	require.NoError(t, err)
	assert.Len(t, secrets, 1)
	assert.Equal(t, "production", secrets[0].Environment)
}
