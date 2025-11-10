package repository

import (
	"context"
	"testing"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// using implementation from parity_repository.go

func TestParityRepository_CreateAndLatest(t *testing.T) {
	r := NewParityRepository()
	c1 := &domain.EnvParityCheck{ProjectID: "p", ScanTimestamp: time.Now().Add(-time.Hour)}
	c2 := &domain.EnvParityCheck{ProjectID: "p", ScanTimestamp: time.Now()}
	require.NoError(t, r.Create(context.Background(), c1))
	require.NoError(t, r.Create(context.Background(), c2))
	latest, err := r.GetLatest(context.Background(), "p")
	require.NoError(t, err)
	assert.Equal(t, c2, latest)
}

func TestParityRepository_GetHistoryLimit(t *testing.T) {
	r := NewParityRepository()
	for i := 0; i < 5; i++ {
		r.Create(context.Background(), &domain.EnvParityCheck{ProjectID: "p", ScanTimestamp: time.Now().Add(time.Duration(-i)*time.Minute)})
	}
	h, err := r.GetHistory(context.Background(), "p", 3)
	require.NoError(t, err)
	assert.Len(t, h, 3)
}
