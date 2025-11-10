package repository

import (
	"context"
	"fmt"
	"time"

	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/storage"
)

type SLORepository interface {
	Create(ctx context.Context, slo *models.SLO) error
	GetByID(ctx context.Context, id string) (*models.SLO, error)
	ListByProject(ctx context.Context, projectID string) ([]*models.SLO, error)
	ListAll(ctx context.Context) ([]*models.SLO, error)
	Update(ctx context.Context, slo *models.SLO) error
	Delete(ctx context.Context, id string) error
	SaveStatus(ctx context.Context, status *models.SLOStatus) error
	GetStatus(ctx context.Context, sloID string) (*models.SLOStatus, error)
	GetHistory(ctx context.Context, sloID string, from, to time.Time) ([]*models.SLOHistory, error)
}

type postgresSLORepo struct {
	db *storage.PostgresDB
}

func NewSLORepository(db *storage.PostgresDB) SLORepository {
	return &postgresSLORepo{db: db}
}

func (r *postgresSLORepo) Create(ctx context.Context, slo *models.SLO) error {
	const q = `INSERT INTO slos (id, project_id, service_name, type, target, window_seconds, query, status, created_at, updated_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,'active',NOW(),NOW())`
	_, err := r.db.DB().ExecContext(ctx, q, slo.ID, slo.ProjectID, slo.ServiceName, string(slo.Type), slo.Target, int64(slo.Window.Seconds()), slo.Query)
	if err != nil {
		return fmt.Errorf("insert slo: %w", err)
	}
	return nil
}

func (r *postgresSLORepo) GetByID(ctx context.Context, id string) (*models.SLO, error) {
	const q = `SELECT id, project_id, service_name, type, target, window_seconds, query, status, created_at, updated_at FROM slos WHERE id = $1`
	row := r.db.DB().QueryRowContext(ctx, q, id)
	var (
		res           models.SLO
		windowSeconds int64
	)
	if err := row.Scan(&res.ID, &res.ProjectID, &res.ServiceName, &res.Type, &res.Target, &windowSeconds, &res.Query, &res.Status, &res.CreatedAt, &res.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get slo: %w", err)
	}
	res.Window = time.Duration(windowSeconds) * time.Second
	return &res, nil
}

func (r *postgresSLORepo) ListByProject(ctx context.Context, projectID string) ([]*models.SLO, error) {
	const q = `SELECT id, project_id, service_name, type, target, window_seconds, query, status, created_at, updated_at FROM slos WHERE project_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.DB().QueryContext(ctx, q, projectID)
	if err != nil {
		return nil, fmt.Errorf("list slos: %w", err)
	}
	defer rows.Close()
	var out []*models.SLO
	for rows.Next() {
		var (
			s  models.SLO
			ws int64
		)
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.ServiceName, &s.Type, &s.Target, &ws, &s.Query, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Window = time.Duration(ws) * time.Second
		out = append(out, &s)
	}
	return out, nil
}

func (r *postgresSLORepo) ListAll(ctx context.Context) ([]*models.SLO, error) {
	const q = `SELECT id, project_id, service_name, type, target, window_seconds, query, status, created_at, updated_at FROM slos WHERE status='active' ORDER BY created_at DESC`
	rows, err := r.db.DB().QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list all slos: %w", err)
	}
	defer rows.Close()
	var out []*models.SLO
	for rows.Next() {
		var (
			s  models.SLO
			ws int64
		)
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.ServiceName, &s.Type, &s.Target, &ws, &s.Query, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Window = time.Duration(ws) * time.Second
		out = append(out, &s)
	}
	return out, nil
}

func (r *postgresSLORepo) Update(ctx context.Context, slo *models.SLO) error {
	const q = `UPDATE slos SET project_id=$2, service_name=$3, type=$4, target=$5, window_seconds=$6, query=$7, status=$8, updated_at=NOW() WHERE id=$1`
	_, err := r.db.DB().ExecContext(ctx, q, slo.ID, slo.ProjectID, slo.ServiceName, string(slo.Type), slo.Target, int64(slo.Window.Seconds()), slo.Query, slo.Status)
	if err != nil {
		return fmt.Errorf("update slo: %w", err)
	}
	return nil
}

func (r *postgresSLORepo) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM slos WHERE id=$1`
	_, err := r.db.DB().ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete slo: %w", err)
	}
	return nil
}

func (r *postgresSLORepo) SaveStatus(ctx context.Context, status *models.SLOStatus) error {
	const q = `INSERT INTO slo_status (slo_id, compliance, burn_rate, remaining_budget, last_calculated)
	VALUES ($1,$2,$3,$4,NOW())
	ON CONFLICT (slo_id) DO UPDATE SET compliance = EXCLUDED.compliance, burn_rate = EXCLUDED.burn_rate, remaining_budget = EXCLUDED.remaining_budget, last_calculated = NOW()`
	_, err := r.db.DB().ExecContext(ctx, q, status.SLOID, status.Compliance, status.BurnRate, status.RemainingBudget)
	if err != nil {
		return fmt.Errorf("upsert slo_status: %w", err)
	}
	return nil
}

func (r *postgresSLORepo) GetStatus(ctx context.Context, sloID string) (*models.SLOStatus, error) {
	const q = `SELECT slo_id, compliance, burn_rate, remaining_budget, last_calculated FROM slo_status WHERE slo_id = $1`
	row := r.db.DB().QueryRowContext(ctx, q, sloID)
	var s models.SLOStatus
	if err := row.Scan(&s.SLOID, &s.Compliance, &s.BurnRate, &s.RemainingBudget, &s.LastCalculated); err != nil {
		return nil, fmt.Errorf("get slo_status: %w", err)
	}
	return &s, nil
}

func (r *postgresSLORepo) GetHistory(ctx context.Context, sloID string, from, to time.Time) ([]*models.SLOHistory, error) {
	const q = `SELECT slo_id, timestamp, compliance, burn_rate FROM slo_history WHERE slo_id = $1 AND timestamp BETWEEN $2 AND $3 ORDER BY timestamp DESC LIMIT 1000`
	rows, err := r.db.DB().QueryContext(ctx, q, sloID, from, to)
	if err != nil {
		return nil, fmt.Errorf("get slo_history: %w", err)
	}
	defer rows.Close()
	var out []*models.SLOHistory
	for rows.Next() {
		var h models.SLOHistory
		if err := rows.Scan(&h.SLOID, &h.Timestamp, &h.Compliance, &h.BurnRate); err != nil {
			return nil, err
		}
		out = append(out, &h)
	}
	return out, nil
}
