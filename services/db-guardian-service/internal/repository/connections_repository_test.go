package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"example.com/lma/db-guardian-service/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestConnections_Create_InsertsAndReturnsID(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewConnectionsRepository(db)
	ctx := context.Background()
	
	conn := &models.DBConnection{
		ProjectID: uuid.New().String(),
		Name:      "primary",
		Driver:    "postgres",
		DSNRef:    "secret/data/db/prod",
		IsDefault: true,
	}

	expectedID := uuid.New().String()
	mock.ExpectQuery(`INSERT INTO db_connections`).
		WithArgs(conn.ProjectID, conn.Name, conn.Driver, conn.DSNRef, conn.IsDefault).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedID))

	// Act
	id, err := repo.Create(ctx, conn)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != expectedID {
		t.Errorf("expected ID %s, got %s", expectedID, id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestConnections_GetDefaultByProject_Found(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewConnectionsRepository(db)
	ctx := context.Background()
	projectID := uuid.New().String()

	expectedConn := &models.DBConnection{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		Name:      "primary",
		Driver:    "postgres",
		DSNRef:    "secret/data/db/prod",
		IsDefault: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "driver", "dsn_ref", "is_default", "created_at", "updated_at"}).
		AddRow(expectedConn.ID, expectedConn.ProjectID, expectedConn.Name, expectedConn.Driver, 
			expectedConn.DSNRef, expectedConn.IsDefault, expectedConn.CreatedAt, expectedConn.UpdatedAt)

	mock.ExpectQuery(`SELECT (.+) FROM db_connections WHERE project_id = \$1 AND is_default = TRUE`).
		WithArgs(projectID).
		WillReturnRows(rows)

	// Act
	conn, err := repo.GetDefaultByProject(ctx, projectID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if conn.ID != expectedConn.ID {
		t.Errorf("expected ID %s, got %s", expectedConn.ID, conn.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestConnections_GetDefaultByProject_NotFound(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewConnectionsRepository(db)
	ctx := context.Background()
	projectID := uuid.New().String()

	mock.ExpectQuery(`SELECT (.+) FROM db_connections WHERE project_id = \$1 AND is_default = TRUE`).
		WithArgs(projectID).
		WillReturnError(sql.ErrNoRows)

	// Act
	_, err = repo.GetDefaultByProject(ctx, projectID)

	// Assert
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestConnections_ListByProject_ReturnsSlice(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewConnectionsRepository(db)
	ctx := context.Background()
	projectID := uuid.New().String()

	rows := sqlmock.NewRows([]string{"id", "project_id", "name", "driver", "dsn_ref", "is_default", "created_at", "updated_at"}).
		AddRow(uuid.New().String(), projectID, "primary", "postgres", "secret/data/db1", true, time.Now(), time.Now()).
		AddRow(uuid.New().String(), projectID, "replica", "postgres", "secret/data/db2", false, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT (.+) FROM db_connections WHERE project_id = \$1 ORDER BY created_at DESC`).
		WithArgs(projectID).
		WillReturnRows(rows)

	// Act
	conns, err := repo.ListByProject(ctx, projectID)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(conns) != 2 {
		t.Errorf("expected 2 connections, got %d", len(conns))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestConnections_Delete_RemovesRow(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	repo := NewConnectionsRepository(db)
	ctx := context.Background()
	id := uuid.New().String()

	mock.ExpectExec(`DELETE FROM db_connections WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Act
	err = repo.Delete(ctx, id)

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
