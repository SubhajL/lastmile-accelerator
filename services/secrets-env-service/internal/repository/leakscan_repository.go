package repository

import (
	"context"
	"strings"

	"example.com/lma/secrets-env-service/internal/domain"
)

type LeakScanRepository struct { data map[string]*domain.ClientLeakScan }

func NewLeakScanRepository() *LeakScanRepository { return &LeakScanRepository{data: make(map[string]*domain.ClientLeakScan)} }

func (r *LeakScanRepository) CreateBatch(ctx context.Context, scans []*domain.ClientLeakScan) error {
	for _, s := range scans { r.data[s.ID] = s }
	return nil
}
func (r *LeakScanRepository) GetBySnapshotID(ctx context.Context, snapshotID, severity string) ([]*domain.ClientLeakScan, error) {
	var out []*domain.ClientLeakScan
	for _, s := range r.data { if s.SnapshotID == snapshotID { if severity==""||strings.EqualFold(severity,s.Severity){ out=append(out,s) } } }
	return out, nil
}
func (r *LeakScanRepository) MarkAsFixed(ctx context.Context, scanID string) error {
	if v, ok := r.data[scanID]; ok { v.Fixed = true }
	return nil
}
