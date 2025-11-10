package repository

import (
	"context"
	"database/sql"
	"fmt"

	"example.com/lma/db-guardian-service/internal/models"
)

type ConnectionsRepository struct {
	db SQLExecutor
}

func NewConnectionsRepository(db SQLExecutor) *ConnectionsRepository {
	return &ConnectionsRepository{db: db}
}

func (r *ConnectionsRepository) Create(ctx context.Context, conn *models.DBConnection) (string, error) {
	if conn == nil {
		return "", ErrInvalidInput
	}
	if conn.ProjectID == "" || conn.Name == "" || conn.Driver == "" || conn.DSNRef == "" {
		return "", ErrInvalidInput
	}

	query := `
		INSERT INTO db_connections (project_id, name, driver, dsn_ref, is_default)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query, 
		conn.ProjectID, conn.Name, conn.Driver, conn.DSNRef, conn.IsDefault,
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("failed to create connection: %w", err)
	}

	return id, nil
}

func (r *ConnectionsRepository) GetDefaultByProject(ctx context.Context, projectID string) (*models.DBConnection, error) {
	if projectID == "" {
		return nil, ErrInvalidInput
	}

	query := `
		SELECT id, project_id, name, driver, dsn_ref, is_default, created_at, updated_at
		FROM db_connections
		WHERE project_id = $1 AND is_default = TRUE
	`

	conn := &models.DBConnection{}
	err := r.db.QueryRowContext(ctx, query, projectID).Scan(
		&conn.ID, &conn.ProjectID, &conn.Name, &conn.Driver, 
		&conn.DSNRef, &conn.IsDefault, &conn.CreatedAt, &conn.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get default connection: %w", err)
	}

	return conn, nil
}

func (r *ConnectionsRepository) ListByProject(ctx context.Context, projectID string) ([]models.DBConnection, error) {
	if projectID == "" {
		return nil, ErrInvalidInput
	}

	query := `
		SELECT id, project_id, name, driver, dsn_ref, is_default, created_at, updated_at
		FROM db_connections
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list connections: %w", err)
	}
	defer rows.Close()

	var connections []models.DBConnection
	for rows.Next() {
		var conn models.DBConnection
		err := rows.Scan(
			&conn.ID, &conn.ProjectID, &conn.Name, &conn.Driver,
			&conn.DSNRef, &conn.IsDefault, &conn.CreatedAt, &conn.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan connection: %w", err)
		}
		connections = append(connections, conn)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating connections: %w", err)
	}

	return connections, nil
}

func (r *ConnectionsRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}

	query := `DELETE FROM db_connections WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete connection: %w", err)
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
