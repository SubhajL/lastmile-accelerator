package service

import (
	"context"
	"testing"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"example.com/lma/secrets-env-service/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareSecretKeys_Sets(t *testing.T) {
	missing, extra := compareSecretKeys([]string{"A","B"}, []string{"B","C"})
	assert.Equal(t, []string{"A"}, missing)
	assert.Equal(t, []string{"C"}, extra)
}

func TestParityService_CheckParity(t *testing.T) {
	secretsRepo := repository.NewSecretsRepository(nil)
	// populate
	_ = secretsRepo.Create(context.Background(), &domain.Secret{ID:"1", ProjectID:"p", Key:"A", Environment:"dev", CreatedAt: time.Now()})
	_ = secretsRepo.Create(context.Background(), &domain.Secret{ID:"2", ProjectID:"p", Key:"B", Environment:"dev", CreatedAt: time.Now()})
	_ = secretsRepo.Create(context.Background(), &domain.Secret{ID:"3", ProjectID:"p", Key:"B", Environment:"prod", CreatedAt: time.Now()})
	_ = secretsRepo.Create(context.Background(), &domain.Secret{ID:"4", ProjectID:"p", Key:"C", Environment:"prod", CreatedAt: time.Now()})

	parityRepo := repository.NewParityRepository()
svc := NewParityService(parityRepo, secretsRepo, nil)

	check, err := svc.CheckParity(context.Background(), "p", "dev", "prod")
	require.NoError(t, err)
	assert.Equal(t, []string{"A"}, check.MissingKeys)
	assert.Equal(t, []string{"C"}, check.ExtraKeys)
}
