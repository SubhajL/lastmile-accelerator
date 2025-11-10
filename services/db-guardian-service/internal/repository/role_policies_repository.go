package repository

import (
	"context"
	"fmt"

	"example.com/lma/db-guardian-service/internal/models"
)

type RolePoliciesRepository struct {
	db SQLExecutor
}

func NewRolePoliciesRepository(db SQLExecutor) *RolePoliciesRepository {
	return &RolePoliciesRepository{db: db}
}

func (r *RolePoliciesRepository) GetByProject(ctx context.Context, projectID string) (*models.RolePolicy, error) {
	if projectID == "" { return nil, ErrInvalidInput }
	query := `SELECT id, project_id, spec_yaml, version, created_at, updated_at FROM role_policies WHERE project_id = $1`
	row := r.db.QueryRowContext(ctx, query, projectID)
	var p models.RolePolicy
	if err := row.Scan(&p.ID, &p.ProjectID, &p.SpecYAML, &p.Version, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, fmt.Errorf("get role policy: %w", err)
	}
	return &p, nil
}

func (r *RolePoliciesRepository) Upsert(ctx context.Context, p *models.RolePolicy) (string, error) {
	if p == nil || p.ProjectID == "" || p.SpecYAML == "" { return "", ErrInvalidInput }
	query := `
	INSERT INTO role_policies (project_id, spec_yaml, version)
	VALUES ($1, $2, COALESCE($3,1))
	ON CONFLICT (project_id)
	DO UPDATE SET spec_yaml = EXCLUDED.spec_yaml, version = role_policies.version + 1, updated_at = NOW()
	RETURNING id, version
	`
	var id string
	var ver int
	if err := r.db.QueryRowContext(ctx, query, p.ProjectID, p.SpecYAML, p.Version).Scan(&id, &ver); err != nil {
		return "", fmt.Errorf("upsert role policy: %w", err)
	}
	p.ID = id; p.Version = ver
	return id, nil
}
