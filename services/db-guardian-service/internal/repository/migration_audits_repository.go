package repository

import (
	"context"
	"fmt"

	"example.com/lma/db-guardian-service/internal/models"
)

type MigrationAuditsRepository struct {
	db SQLExecutor
}

func NewMigrationAuditsRepository(db SQLExecutor) *MigrationAuditsRepository {
	return &MigrationAuditsRepository{db: db}
}

func (r *MigrationAuditsRepository) Record(ctx context.Context, audit *models.MigrationAudit) (string, error) {
	if audit == nil {
		return "", ErrInvalidInput
	}
	if audit.ProjectID == "" || audit.MigrationName == "" || audit.Status == "" {
		return "", ErrInvalidInput
	}

	query := `
		INSERT INTO migration_audits (project_id, migration_name, status, findings_json)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query,
		audit.ProjectID, audit.MigrationName, audit.Status, audit.FindingsJSON,
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("failed to record migration audit: %w", err)
	}

	return id, nil
}

func (r *MigrationAuditsRepository) ListByProject(ctx context.Context, projectID string, limit int) ([]models.MigrationAudit, error) {
	if projectID == "" {
		return nil, ErrInvalidInput
	}
	if limit <= 0 {
		limit = 100 // default limit
	}

	query := `
		SELECT id, project_id, migration_name, status, findings_json, created_at
		FROM migration_audits
		WHERE project_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, projectID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list migration audits: %w", err)
	}
    defer func(){ _ = rows.Close() }()

	var audits []models.MigrationAudit
	for rows.Next() {
		var audit models.MigrationAudit
		err := rows.Scan(
			&audit.ID, &audit.ProjectID, &audit.MigrationName,
			&audit.Status, &audit.FindingsJSON, &audit.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan migration audit: %w", err)
		}
		audits = append(audits, audit)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration audits: %w", err)
	}

	return audits, nil
}
