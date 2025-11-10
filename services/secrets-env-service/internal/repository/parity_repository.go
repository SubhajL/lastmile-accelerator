package repository

import (
	"context"
	"sort"

	"example.com/lma/secrets-env-service/internal/domain"
	"github.com/stretchr/testify/assert"
)

type ParityRepository struct {
	data map[string][]*domain.EnvParityCheck // projectID -> checks
}

func NewParityRepository() *ParityRepository {
	return &ParityRepository{data: make(map[string][]*domain.EnvParityCheck)}
}

func (r *ParityRepository) Create(ctx context.Context, check *domain.EnvParityCheck) error {
	r.data[check.ProjectID] = append(r.data[check.ProjectID], check)
	return nil
}

func (r *ParityRepository) GetLatest(ctx context.Context, projectID string) (*domain.EnvParityCheck, error) {
	arr := r.data[projectID]
	if len(arr) == 0 { return nil, assert.AnError }
	sort.Slice(arr, func(i,j int) bool { return arr[i].ScanTimestamp.After(arr[j].ScanTimestamp) })
	return arr[0], nil
}

func (r *ParityRepository) GetHistory(ctx context.Context, projectID string, limit int) ([]*domain.EnvParityCheck, error) {
	arr := r.data[projectID]
	sort.Slice(arr, func(i,j int) bool { return arr[i].ScanTimestamp.After(arr[j].ScanTimestamp) })
	if limit > 0 && len(arr) > limit { arr = arr[:limit] }
	return arr, nil
}
