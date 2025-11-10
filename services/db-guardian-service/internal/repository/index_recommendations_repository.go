package repository

import (
	"context"
	"fmt"

	"example.com/lma/db-guardian-service/internal/models"
	"github.com/lib/pq"
)

type IndexRecommendationsRepository struct {
	db SQLExecutor
}

func NewIndexRecommendationsRepository(db SQLExecutor) *IndexRecommendationsRepository {
	return &IndexRecommendationsRepository{db: db}
}

func (r *IndexRecommendationsRepository) Upsert(ctx context.Context, rec *models.IndexRecommendation) (string, error) {
	if rec == nil {
		return "", ErrInvalidInput
	}
	if rec.ProjectID == "" || rec.TableName == "" || len(rec.ColumnNames) == 0 {
		return "", ErrInvalidInput
	}

	query := `
		INSERT INTO index_recommendations (project_id, table_name, column_names, reason, benefit_score, applied)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (project_id, table_name, column_names)
		DO UPDATE SET
			reason = EXCLUDED.reason,
			benefit_score = EXCLUDED.benefit_score,
			applied = EXCLUDED.applied,
			updated_at = NOW()
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query,
		rec.ProjectID, rec.TableName, pq.Array(rec.ColumnNames),
		rec.Reason, rec.BenefitScore, rec.Applied,
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("failed to upsert index recommendation: %w", err)
	}

	return id, nil
}

func (r *IndexRecommendationsRepository) ListByProject(ctx context.Context, projectID string, onlyUnapplied bool) ([]models.IndexRecommendation, error) {
	if projectID == "" {
		return nil, ErrInvalidInput
	}

	query := `
		SELECT id, project_id, table_name, column_names, reason, benefit_score, applied, created_at, updated_at
		FROM index_recommendations
		WHERE project_id = $1
	`

	args := []interface{}{projectID}
	if onlyUnapplied {
		query += ` AND applied = FALSE`
	}
	query += ` ORDER BY benefit_score DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list index recommendations: %w", err)
	}
	defer rows.Close()

	var recommendations []models.IndexRecommendation
	for rows.Next() {
		var rec models.IndexRecommendation
		var columnNames pq.StringArray
		err := rows.Scan(
			&rec.ID, &rec.ProjectID, &rec.TableName, &columnNames,
			&rec.Reason, &rec.BenefitScore, &rec.Applied, &rec.CreatedAt, &rec.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan index recommendation: %w", err)
		}
		rec.ColumnNames = []string(columnNames)
		recommendations = append(recommendations, rec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating index recommendations: %w", err)
	}

	return recommendations, nil
}

func (r *IndexRecommendationsRepository) MarkApplied(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}

	query := `
		UPDATE index_recommendations
		SET applied = TRUE, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark recommendation as applied: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}
