package service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"example.com/lma/db-guardian-service/internal/models"
	"github.com/DATA-DOG/go-sqlmock"
)

type fakeConnGetter struct {
	conn *models.DBConnection
	err  error
}

func (f *fakeConnGetter) GetDefaultConnection(ctx context.Context, projectID string) (*models.DBConnection, error) {
	return f.conn, f.err
}

type fakeVault struct {
	dsn string
	err error
}

func (f *fakeVault) GetProjectDSN(ctx context.Context, path string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.dsn, nil
}

func TestProjectResolver_ResolveInspector_HappyPath(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	mock.ExpectClose()

	r := NewProjectResolver(
		&fakeConnGetter{conn: &models.DBConnection{ProjectID: "p1", Driver: "postgres", DSNRef: "secret/data/dbs/p1/default"}},
		&fakeVault{dsn: "postgres://example"},
		func(ctx context.Context, dsn string) (*sql.DB, error) { return db, nil },
	)

	// Act
	inspector, cleanup, err := r.ResolveInspector(context.Background(), "p1")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if inspector == nil {
		t.Fatal("expected inspector")
	}
	if cleanup == nil {
		t.Fatal("expected cleanup")
	}
	if err := cleanup(); err != nil {
		t.Fatalf("cleanup: %v", err)
	}
}

func TestProjectResolver_ResolveInspector_UnsupportedDriver(t *testing.T) {
	r := NewProjectResolver(
		&fakeConnGetter{conn: &models.DBConnection{ProjectID: "p1", Driver: "mysql", DSNRef: "secret/data/dbs/p1/default"}},
		&fakeVault{dsn: "mysql://example"},
		func(ctx context.Context, dsn string) (*sql.DB, error) { return nil, nil },
	)

	_, _, err := r.ResolveInspector(context.Background(), "p1")
	if err == nil || err.Error() != "unsupported driver: mysql" {
		t.Fatalf("expected unsupported driver error, got %v", err)
	}
}

func TestProjectResolver_ResolveInspector_VaultError(t *testing.T) {
	r := NewProjectResolver(
		&fakeConnGetter{conn: &models.DBConnection{ProjectID: "p1", Driver: "postgres", DSNRef: "secret/data/dbs/p1/default"}},
		&fakeVault{err: fmt.Errorf("boom")},
		func(ctx context.Context, dsn string) (*sql.DB, error) { return nil, nil },
	)

	_, _, err := r.ResolveInspector(context.Background(), "p1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProjectResolver_ResolveInspector_EmptyProjectID(t *testing.T) {
	r := NewProjectResolver(
		&fakeConnGetter{conn: &models.DBConnection{ProjectID: "p1", Driver: "postgres", DSNRef: "secret/data/dbs/p1/default"}},
		&fakeVault{dsn: "postgres://example"},
		func(ctx context.Context, dsn string) (*sql.DB, error) { return nil, nil },
	)

	_, _, err := r.ResolveInspector(context.Background(), "")
	if err == nil || err.Error() != "projectID is required" {
		t.Fatalf("expected projectID is required, got %v", err)
	}
}

func TestProjectResolver_ResolveInspector_NilConnection(t *testing.T) {
	r := NewProjectResolver(
		&fakeConnGetter{conn: nil},
		&fakeVault{dsn: "postgres://example"},
		func(ctx context.Context, dsn string) (*sql.DB, error) { return nil, nil },
	)

	_, _, err := r.ResolveInspector(context.Background(), "p1")
	if err == nil || err.Error() != "default connection not found" {
		t.Fatalf("expected default connection not found, got %v", err)
	}
}

func TestProjectResolver_ResolveInspector_EmptyDSNRef(t *testing.T) {
	r := NewProjectResolver(
		&fakeConnGetter{conn: &models.DBConnection{ProjectID: "p1", Driver: "postgres", DSNRef: ""}},
		&fakeVault{dsn: "postgres://example"},
		func(ctx context.Context, dsn string) (*sql.DB, error) { return nil, nil },
	)

	_, _, err := r.ResolveInspector(context.Background(), "p1")
	if err == nil || err.Error() != "dsn_ref is required" {
		t.Fatalf("expected dsn_ref is required, got %v", err)
	}
}

func TestProjectResolver_ResolveInspector_DBOpenFailure(t *testing.T) {
	r := NewProjectResolver(
		&fakeConnGetter{conn: &models.DBConnection{ProjectID: "p1", Driver: "postgres", DSNRef: "secret/data/dbs/p1/default"}},
		&fakeVault{dsn: "postgres://example"},
		func(ctx context.Context, dsn string) (*sql.DB, error) { return nil, fmt.Errorf("boom") },
	)

	_, _, err := r.ResolveInspector(context.Background(), "p1")
	if err == nil {
		t.Fatal("expected error")
	}
}
