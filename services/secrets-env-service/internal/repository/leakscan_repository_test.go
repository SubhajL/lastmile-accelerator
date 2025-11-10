package repository

import (
	"context"
	"testing"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// using implementation from leakscan_repository.go

func TestLeakScanRepository_CreateAndGet(t *testing.T) {
	r := NewLeakScanRepository()
	scans := []*domain.ClientLeakScan{{ID:"1", SnapshotID:"s", Severity:"high", CreatedAt: time.Now()}}
	require.NoError(t, r.CreateBatch(context.Background(), scans))
	got, err := r.GetBySnapshotID(context.Background(), "s", "")
	require.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestLeakScanRepository_FilterSeverity(t *testing.T) {
	r := NewLeakScanRepository()
	r.CreateBatch(context.Background(), []*domain.ClientLeakScan{{ID:"1", SnapshotID:"s", Severity:"high"},{ID:"2", SnapshotID:"s", Severity:"low"}})
	high, _ := r.GetBySnapshotID(context.Background(), "s", "high")
	assert.Len(t, high, 1)
}

func TestLeakScanRepository_MarkAsFixed(t *testing.T) {
	r := NewLeakScanRepository()
	s := &domain.ClientLeakScan{ID:"1", SnapshotID:"s", Severity:"high"}
	r.CreateBatch(context.Background(), []*domain.ClientLeakScan{s})
	require.NoError(t, r.MarkAsFixed(context.Background(), "1"))
	assert.True(t, r.data["1"].Fixed)
}
