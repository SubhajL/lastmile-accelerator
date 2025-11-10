package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
)

// Postgres implementation for parity repository

type ParityRepositoryPG struct{ db *sql.DB }

func NewParityRepositoryPostgres(db *sql.DB) *ParityRepositoryPG { return &ParityRepositoryPG{db: db} }

func (r *ParityRepositoryPG) Create(ctx context.Context, check *domain.EnvParityCheck) error {
	missing, _ := json.Marshal(check.MissingKeys)
	extra, _ := json.Marshal(check.ExtraKeys)
	hasDrift := len(check.MissingKeys) > 0 || len(check.ExtraKeys) > 0
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO env_parity_checks (project_id, scan_timestamp, missing_keys, extra_keys, has_drift) VALUES ($1,$2,$3,$4,$5)`,
		check.ProjectID, check.ScanTimestamp, string(missing), string(extra), hasDrift,
	)
	return err
}

func (r *ParityRepositoryPG) GetLatest(ctx context.Context, projectID string) (*domain.EnvParityCheck, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT project_id, scan_timestamp, missing_keys, extra_keys, has_drift FROM env_parity_checks WHERE project_id=$1 ORDER BY scan_timestamp DESC LIMIT 1`, projectID,
	)
	var (
		proj string
		ts time.Time
		missingJSON, extraJSON string
		hasDrift bool
	)
	if err := row.Scan(&proj, &ts, &missingJSON, &extraJSON, &hasDrift); err != nil {
		return nil, err
	}
	var missing, extra []string
	_ = json.Unmarshal([]byte(missingJSON), &missing)
	_ = json.Unmarshal([]byte(extraJSON), &extra)
	return &domain.EnvParityCheck{ProjectID: proj, ScanTimestamp: ts, MissingKeys: missing, ExtraKeys: extra}, nil
}

func (r *ParityRepositoryPG) GetHistory(ctx context.Context, projectID string, limit int) ([]*domain.EnvParityCheck, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT project_id, scan_timestamp, missing_keys, extra_keys, has_drift FROM env_parity_checks WHERE project_id=$1 ORDER BY scan_timestamp DESC LIMIT $2`, projectID, limit,
	)
	if err != nil { return nil, err }
	defer rows.Close()
	var out []*domain.EnvParityCheck
	for rows.Next() {
		var (
			proj string
			ts time.Time
			missingJSON, extraJSON string
			hasDrift bool
		)
		if err := rows.Scan(&proj, &ts, &missingJSON, &extraJSON, &hasDrift); err != nil { return nil, err }
		var missing, extra []string
		_ = json.Unmarshal([]byte(missingJSON), &missing)
		_ = json.Unmarshal([]byte(extraJSON), &extra)
		out = append(out, &domain.EnvParityCheck{ProjectID: proj, ScanTimestamp: ts, MissingKeys: missing, ExtraKeys: extra})
	}
	if len(out) == 0 {
		return out, nil
	}
	return out, nil
}
