package service

import (
	"context"
	"database/sql"
	"fmt"

	"example.com/lma/db-guardian-service/internal/database"
	"example.com/lma/db-guardian-service/internal/models"
	"example.com/lma/db-guardian-service/internal/repository"
)

// ConnectionService provides high-level operations for managing DB connections per project.
type ConnectionService struct {
	db   *sql.DB
	repo *repository.ConnectionsRepository
}

func NewConnectionService(db *sql.DB) *ConnectionService {
	return &ConnectionService{
		db:   db,
		repo: repository.NewConnectionsRepository(db),
	}
}

// RegisterConnection creates a new DB connection; if makeDefault is true, it becomes the project's default.
func (s *ConnectionService) RegisterConnection(ctx context.Context, conn *models.DBConnection, makeDefault bool) (string, error) {
	if conn == nil {
		return "", repository.ErrInvalidInput
	}
	if !makeDefault {
		return s.repo.Create(ctx, conn)
	}

	var newID string
	err := database.WithTx(ctx, s.db, func(ctx context.Context, tx database.Tx) error {
		// Unset existing defaults for project
		if _, err := tx.ExecContext(ctx, `UPDATE db_connections SET is_default = FALSE WHERE project_id = $1`, conn.ProjectID); err != nil {
			return fmt.Errorf("unset previous defaults: %w", err)
		}

		// Create new connection as default
		repoTx := repository.NewConnectionsRepository(tx)
		conn.IsDefault = true
		id, err := repoTx.Create(ctx, conn)
		if err != nil {
			return err
		}
		newID = id
		return nil
	})
	if err != nil {
		return "", err
	}
	return newID, nil
}

// DeleteConnection removes a connection by ID.
func (s *ConnectionService) DeleteConnection(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// GetDefaultConnection returns the default connection for a project.
func (s *ConnectionService) GetDefaultConnection(ctx context.Context, projectID string) (*models.DBConnection, error) {
	return s.repo.GetDefaultByProject(ctx, projectID)
}

// ListConnections returns all connections for a project.
func (s *ConnectionService) ListConnections(ctx context.Context, projectID string) ([]models.DBConnection, error) {
	return s.repo.ListByProject(ctx, projectID)
}
