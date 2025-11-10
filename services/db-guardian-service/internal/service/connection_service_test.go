package service

import (
	"context"
	"regexp"
	"testing"

	"example.com/lma/db-guardian-service/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestConnectionService_RegisterConnection_SetsDefault(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	svc := NewConnectionService(db)
	ctx := context.Background()
	conn := &models.DBConnection{
		ProjectID: "proj-1",
		Name:      "primary",
		Driver:    "postgres",
		DSNRef:    "secret/data/dbs/proj-1/primary",
	}

	// Expect transaction begin
	mock.ExpectBegin()
	// Expect unset previous defaults
	mock.ExpectExec(regexp.QuoteMeta("UPDATE db_connections SET is_default = FALSE WHERE project_id = $1")).
		WithArgs("proj-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Expect insert returning id
	rows := sqlmock.NewRows([]string{"id"}).AddRow("conn-123")
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO db_connections (project_id, name, driver, dsn_ref, is_default) VALUES ($1, $2, $3, $4, $5) RETURNING id")).
		WithArgs("proj-1", "primary", "postgres", "secret/data/dbs/proj-1/primary", true).
		WillReturnRows(rows)
	// Expect commit
	mock.ExpectCommit()

	// Act
	id, err := svc.RegisterConnection(ctx, conn, true)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != "conn-123" {
		t.Errorf("expected id 'conn-123', got '%s'", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestConnectionService_DeleteConnection_Succeeds(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	svc := NewConnectionService(db)
	ctx := context.Background()

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM db_connections WHERE id = $1")).
		WithArgs("conn-123").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Act
	err = svc.DeleteConnection(ctx, "conn-123")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
