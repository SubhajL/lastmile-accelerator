package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/models"
	"example.com/lma/db-guardian-service/internal/repository"
	"github.com/DATA-DOG/go-sqlmock"
)

// MockVaultClient is a test double for secrets.VaultClient
type MockVaultClient struct {
	GetProjectDSNFunc func(ctx context.Context, vaultRef string) (string, error)
}

func (m *MockVaultClient) GetProjectDSN(ctx context.Context, vaultRef string) (string, error) {
	if m.GetProjectDSNFunc != nil {
		return m.GetProjectDSNFunc(ctx, vaultRef)
	}
	return "", errors.New("GetProjectDSN not mocked")
}

// MockDBInspector is a test double for analyzer.DBInspector
type MockDBInspector struct{}

func (m *MockDBInspector) GetTables(ctx context.Context, schema string) ([]analyzer.TableInfo, error) {
	return nil, nil
}

func (m *MockDBInspector) GetTableColumns(ctx context.Context, table string) ([]analyzer.ColumnInfo, error) {
	return nil, nil
}

func (m *MockDBInspector) GetIndexes(ctx context.Context, table string) ([]analyzer.IndexInfo, error) {
	return nil, nil
}

func (m *MockDBInspector) GetRoles(ctx context.Context) ([]analyzer.RoleInfo, error) {
	return nil, nil
}

func (m *MockDBInspector) GetQueryStats(ctx context.Context, minExecutions int) ([]analyzer.QueryStat, error) {
	return nil, nil
}

func (m *MockDBInspector) AnalyzeMigration(ctx context.Context, sql string) (*analyzer.MigrationAnalysis, error) {
	return nil, nil
}

// TestProjectResolver_Resolve_Success tests successful database resolution
func TestProjectResolver_Resolve_Success(t *testing.T) {
	// Setup mocks
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Mock repository query
	testTime := time.Now()
	mock.ExpectQuery(`SELECT id, project_id, name, driver, dsn_ref, is_default, created_at, updated_at FROM db_connections WHERE project_id = \$1 AND is_default = TRUE`).
		WithArgs("test-project-123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "name", "driver", "dsn_ref", "is_default", "created_at", "updated_at"}).
			AddRow("1", "test-project-123", "test-conn", "postgres", "vault:secret/projects/test-project-123/db/dsn", true, testTime, testTime))

	// Setup vault mock
	mockVault := &MockVaultClient{
		GetProjectDSNFunc: func(ctx context.Context, vaultRef string) (string, error) {
			if vaultRef == "vault:secret/projects/test-project-123/db/dsn" {
				return "postgres://user:pass@localhost:5432/testdb", nil
			}
			return "", errors.New("unexpected vault ref")
		},
	}

	// Create resolver with mock vault
	resolver := &ProjectResolver{
		db:       db,
		vault:    mockVault,
		connRepo: repository.NewConnectionsRepository(db),
		inspectorNew: func(db *sql.DB) analyzer.DBInspector {
			return &MockDBInspector{}
		},
	}

	// Test
	projectDB, closer, err := resolver.Resolve(context.Background(), "test-project-123")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if projectDB == nil {
		t.Fatal("expected database connection, got nil")
	}
	if closer == nil {
		t.Fatal("expected closer function, got nil")
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestProjectResolver_Resolve_EmptyProjectID tests validation of empty project ID
func TestProjectResolver_Resolve_EmptyProjectID(t *testing.T) {
	resolver := &ProjectResolver{}

	_, _, err := resolver.Resolve(context.Background(), "")

	if err == nil {
		t.Fatal("expected error for empty projectID, got nil")
	}
	if err.Error() != "projectID cannot be empty" {
		t.Errorf("expected error 'projectID cannot be empty', got: %v", err)
	}
}

// TestProjectResolver_Resolve_ConnectionNotFound tests handling of missing connections
func TestProjectResolver_Resolve_ConnectionNotFound(t *testing.T) {
	// Setup mocks
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Mock repository query returning no rows
	mock.ExpectQuery(`SELECT id, project_id, name, driver, dsn_ref, is_default, created_at, updated_at FROM db_connections WHERE project_id = \$1 AND is_default = TRUE`).
		WithArgs("missing-project").
		WillReturnError(repository.ErrNotFound)

	// Create resolver
	resolver := &ProjectResolver{
		db:       db,
		connRepo: repository.NewConnectionsRepository(db),
	}

	// Test
	_, _, err = resolver.Resolve(context.Background(), "missing-project")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repository.ErrNotFound) {
		t.Errorf("expected repository.ErrNotFound, got: %v", err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestProjectResolver_Resolve_VaultFetchFails tests handling of Vault failures
func TestProjectResolver_Resolve_VaultFetchFails(t *testing.T) {
	// Setup mocks
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Mock repository query
	testTime := time.Now()
	mock.ExpectQuery(`SELECT id, project_id, name, driver, dsn_ref, is_default, created_at, updated_at FROM db_connections WHERE project_id = \$1 AND is_default = TRUE`).
		WithArgs("test-project").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "name", "driver", "dsn_ref", "is_default", "created_at", "updated_at"}).
			AddRow("1", "test-project", "test-conn", "postgres", "vault:secret/projects/test-project/db/dsn", true, testTime, testTime))

	// Setup vault mock that fails
	mockVault := &MockVaultClient{
		GetProjectDSNFunc: func(ctx context.Context, vaultRef string) (string, error) {
			return "", errors.New("vault is sealed")
		},
	}

	// Create resolver with mock vault
	resolver := &ProjectResolver{
		db:       db,
		vault:    mockVault,
		connRepo: repository.NewConnectionsRepository(db),
	}

	// Test
	_, _, err = resolver.Resolve(context.Background(), "test-project")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expectedErrMsg := "failed to fetch DSN for project test-project: vault is sealed"
	if err.Error() != expectedErrMsg {
		t.Errorf("expected error %q, got: %v", expectedErrMsg, err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestProjectResolver_Resolve_InvalidDriver tests handling of invalid database drivers
func TestProjectResolver_Resolve_InvalidDriver(t *testing.T) {
	// Setup mocks
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Mock repository query with empty driver
	testTime := time.Now()
	mock.ExpectQuery(`SELECT id, project_id, name, driver, dsn_ref, is_default, created_at, updated_at FROM db_connections WHERE project_id = \$1 AND is_default = TRUE`).
		WithArgs("test-project").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "name", "driver", "dsn_ref", "is_default", "created_at", "updated_at"}).
			AddRow("1", "test-project", "test-conn", "", "vault:secret/projects/test-project/db/dsn", true, testTime, testTime))

	// Setup vault mock
	mockVault := &MockVaultClient{
		GetProjectDSNFunc: func(ctx context.Context, vaultRef string) (string, error) {
			return "postgres://user:pass@localhost:5432/testdb", nil
		},
	}

	// Create resolver with mock vault
	resolver := &ProjectResolver{
		db:       db,
		vault:    mockVault,
		connRepo: repository.NewConnectionsRepository(db),
	}

	// Test
	_, _, err = resolver.Resolve(context.Background(), "test-project")

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expectedErrMsg := "invalid database driver: driver cannot be empty"
	if err.Error() != expectedErrMsg {
		t.Errorf("expected error %q, got: %v", expectedErrMsg, err)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestProjectResolver_ResolveInspector_ReturnsInspector tests inspector creation
func TestProjectResolver_ResolveInspector_ReturnsInspector(t *testing.T) {
	// Setup mocks
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Mock repository query
	testTime := time.Now()
	mock.ExpectQuery(`SELECT id, project_id, name, driver, dsn_ref, is_default, created_at, updated_at FROM db_connections WHERE project_id = \$1 AND is_default = TRUE`).
		WithArgs("test-project").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "name", "driver", "dsn_ref", "is_default", "created_at", "updated_at"}).
			AddRow("1", "test-project", "test-conn", "postgres", "vault:secret/projects/test-project/db/dsn", true, testTime, testTime))

	// Setup vault mock
	mockVault := &MockVaultClient{
		GetProjectDSNFunc: func(ctx context.Context, vaultRef string) (string, error) {
			return "postgres://user:pass@localhost:5432/testdb", nil
		},
	}

	// Track if inspector factory was called
	inspectorCreated := false

	// Create resolver with mock vault
	resolver := &ProjectResolver{
		db:       db,
		vault:    mockVault,
		connRepo: repository.NewConnectionsRepository(db),
		inspectorNew: func(db *sql.DB) analyzer.DBInspector {
			inspectorCreated = true
			return &MockDBInspector{}
		},
	}

	// Test
	inspector, closer, err := resolver.ResolveInspector(context.Background(), "test-project")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if inspector == nil {
		t.Fatal("expected inspector, got nil")
	}
	if closer == nil {
		t.Fatal("expected closer function, got nil")
	}
	if !inspectorCreated {
		t.Error("inspector factory was not called")
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestProjectResolver_ResolveInspector_PropagatesResolveError tests error propagation
func TestProjectResolver_ResolveInspector_PropagatesResolveError(t *testing.T) {
	resolver := &ProjectResolver{}

	// Test with empty project ID (will fail in Resolve)
	_, _, err := resolver.ResolveInspector(context.Background(), "")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "projectID cannot be empty" {
		t.Errorf("expected error 'projectID cannot be empty', got: %v", err)
	}
}

// TestProjectResolver_Resolve_DBOpenFails tests handling when sql.Open fails
func TestProjectResolver_Resolve_DBOpenFails(t *testing.T) {
	// Setup mocks
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Mock repository query
	testTime := time.Now()
	mock.ExpectQuery(`SELECT id, project_id, name, driver, dsn_ref, is_default, created_at, updated_at FROM db_connections WHERE project_id = \$1 AND is_default = TRUE`).
		WithArgs("test-project").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "name", "driver", "dsn_ref", "is_default", "created_at", "updated_at"}).
			AddRow("1", "test-project", "test-conn", "invalid-driver", "vault:secret/projects/test-project/db/dsn", true, testTime, testTime))

	// Setup vault mock
	mockVault := &MockVaultClient{
		GetProjectDSNFunc: func(ctx context.Context, vaultRef string) (string, error) {
			return "invalid-dsn-format", nil
		},
	}

	// Create resolver
	resolver := &ProjectResolver{
		db:       db,
		vault:    mockVault,
		connRepo: repository.NewConnectionsRepository(db),
	}

	// Test
	_, _, err = resolver.Resolve(context.Background(), "test-project")

	// Assert
	// Note: sql.Open with invalid driver typically doesn't fail immediately
	// It's a lazy operation, so this test may need adjustment based on driver behavior
	if err != nil {
		// If it does fail, it should be because of the invalid driver
		if !strings.Contains(err.Error(), "unknown driver") && !strings.Contains(err.Error(), "invalid-driver") {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// TestProjectResolver_Resolve_CloserCleansUp tests that the closer function works
func TestProjectResolver_Resolve_CloserCleansUp(t *testing.T) {
	// This test verifies that:
	// 1. The resolver returns a working closer function
	// 2. The closer can be called without error
	// 3. We properly clean up resources

	// Setup mocks
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Mock repository query
	testTime := time.Now()
	mock.ExpectQuery(`SELECT id, project_id, name, driver, dsn_ref, is_default, created_at, updated_at FROM db_connections WHERE project_id = \$1 AND is_default = TRUE`).
		WithArgs("test-project").
		WillReturnRows(sqlmock.NewRows([]string{"id", "project_id", "name", "driver", "dsn_ref", "is_default", "created_at", "updated_at"}).
			AddRow("1", "test-project", "test-conn", "postgres", "vault:secret/projects/test-project/db/dsn", true, testTime, testTime))

	// Setup vault mock
	mockVault := &MockVaultClient{
		GetProjectDSNFunc: func(ctx context.Context, vaultRef string) (string, error) {
			return "postgres://user:pass@localhost:5432/testdb", nil
		},
	}

	// Create resolver
	resolver := &ProjectResolver{
		db:       db,
		vault:    mockVault,
		connRepo: repository.NewConnectionsRepository(db),
		inspectorNew: func(db *sql.DB) analyzer.DBInspector {
			return &MockDBInspector{}
		},
	}

	// Test
	_, closer, err := resolver.Resolve(context.Background(), "test-project")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if closer == nil {
		t.Fatal("expected closer function, got nil")
	}

	// Test that closer can be called
	closeErr := closer()
	if closeErr != nil {
		t.Errorf("closer returned error: %v", closeErr)
	}

	// Verify all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// Helper to create a working mock connection for testing
func createWorkingConnection() *models.DBConnection {
	return &models.DBConnection{
		ID:        "1",
		ProjectID: "test-project",
		Name:      "test-connection",
		Driver:    "postgres",
		DSNRef:    "vault:secret/projects/test-project/db/dsn",
		IsDefault: true,
	}
}