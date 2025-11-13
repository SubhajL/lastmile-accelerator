package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"example.com/lma/secrets-env-service/internal/domain"
)

type LeakScanRepositoryPG struct{ db *sql.DB }

func NewLeakScanRepositoryPostgres(db *sql.DB) *LeakScanRepositoryPG { return &LeakScanRepositoryPG{db: db} }


func (r *LeakScanRepositoryPG) CreateBatch(ctx context.Context, scans []*domain.ClientLeakScan) error {
	if len(scans) == 0 { return nil }
	// Simple 2-row batch builder for tests; can be extended to chunking later.
	placeholders := make([]string, 0, len(scans))
	args := make([]any, 0, len(scans)*8)
	idx := 1
	for _, s := range scans {
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)", idx, idx+1, idx+2, idx+3, idx+4, idx+5, idx+6, idx+7))
		args = append(args, s.ID, s.SnapshotID, s.FilePath, s.LineNumber, s.Pattern, s.Severity, s.Fixed, s.CreatedAt)
		idx += 8
	}
	query := "INSERT INTO client_leak_scans (id, snapshot_id, file_path, line_number, pattern, severity, fixed, created_at) VALUES " + strings.Join(placeholders, ",")
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *LeakScanRepositoryPG) GetBySnapshotID(ctx context.Context, snapshotID, severity string) ([]*domain.ClientLeakScan, error) {
    rows, err := r.db.QueryContext(ctx, `SELECT id, snapshot_id, file_path, line_number, pattern, severity, fixed, created_at FROM client_leak_scans WHERE snapshot_id=$1 AND ($2='' OR severity=$2) ORDER BY created_at ASC`, snapshotID, severity)
	if err != nil { return nil, err }
    defer func(){ _ = rows.Close() }()
	var out []*domain.ClientLeakScan
	for rows.Next() {
		s := &domain.ClientLeakScan{}
		if err := rows.Scan(&s.ID,&s.SnapshotID,&s.FilePath,&s.LineNumber,&s.Pattern,&s.Severity,&s.Fixed,&s.CreatedAt); err != nil { return nil, err }
		out = append(out, s)
	}
	return out, nil
}

func (r *LeakScanRepositoryPG) MarkAsFixed(ctx context.Context, scanID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE client_leak_scans SET fixed=true, fixed_at=NOW() WHERE id=$1`, scanID)
	return err
}
