package service

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"github.com/DATA-DOG/go-sqlmock"
)

type fakeInspector struct{}

func (f *fakeInspector) GetTables(ctx context.Context, schema string) ([]analyzer.TableInfo, error) {
	return []analyzer.TableInfo{{Schema: "public", Name: "users", SizeBytes: 64 * 1024 * 1024}}, nil
}
func (f *fakeInspector) GetTableColumns(ctx context.Context, table string) ([]analyzer.ColumnInfo, error) {
	return nil, nil
}
func (f *fakeInspector) GetIndexes(ctx context.Context, table string) ([]analyzer.IndexInfo, error) {
	return []analyzer.IndexInfo{}, nil
}
func (f *fakeInspector) GetRoles(ctx context.Context) ([]analyzer.RoleInfo, error) {
	return []analyzer.RoleInfo{{Name: "app", IsSuperuser: true, CanLogin: true}}, nil
}
func (f *fakeInspector) GetQueryStats(ctx context.Context, minExecutions int) ([]analyzer.QueryStat, error) {
	return []analyzer.QueryStat{{Query: "SELECT * FROM users WHERE email = $1", Calls: 500}}, nil
}
func (f *fakeInspector) AnalyzeMigration(ctx context.Context, sql string) (*analyzer.MigrationAnalysis, error) {
	return &analyzer.MigrationAnalysis{SQL: sql, HasBreaking: true, Operations: []analyzer.SQLOperation{{Type: "ALTER", ObjectType: "COLUMN", ObjectName: "users", Details: map[string]string{"operation": "DROP", "is_breaking": "true"}}}}, nil
}

// MockProjectResolver for testing
type MockProjectResolver struct {
	ResolveInspectorFunc func(ctx context.Context, projectID string) (analyzer.DBInspector, func() error, error)
}

func (m *MockProjectResolver) ResolveInspector(ctx context.Context, projectID string) (analyzer.DBInspector, func() error, error) {
	if m.ResolveInspectorFunc != nil {
		return m.ResolveInspectorFunc(ctx, projectID)
	}
	return nil, nil, nil
}

// errorInspector is a fake inspector that returns errors for testing
type errorInspector struct{}

func (e *errorInspector) GetTables(ctx context.Context, schema string) ([]analyzer.TableInfo, error) {
	return nil, nil
}
func (e *errorInspector) GetTableColumns(ctx context.Context, table string) ([]analyzer.ColumnInfo, error) {
	return nil, nil
}
func (e *errorInspector) GetIndexes(ctx context.Context, table string) ([]analyzer.IndexInfo, error) {
	return nil, nil
}
func (e *errorInspector) GetRoles(ctx context.Context) ([]analyzer.RoleInfo, error) {
	return nil, errors.New("forced error for testing")
}
func (e *errorInspector) GetQueryStats(ctx context.Context, minExecutions int) ([]analyzer.QueryStat, error) {
	return nil, nil
}
func (e *errorInspector) AnalyzeMigration(ctx context.Context, sql string) (*analyzer.MigrationAnalysis, error) {
	return nil, nil
}

func TestAnalysisService_RunProjectAnalysis_Success(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Track if closer was called
	closerCalled := false
	mockResolver := &MockProjectResolver{
		ResolveInspectorFunc: func(ctx context.Context, projectID string) (analyzer.DBInspector, func() error, error) {
			if projectID != "proj-1" {
				t.Errorf("expected projectID 'proj-1', got '%s'", projectID)
			}
			return &fakeInspector{}, func() error {
				closerCalled = true
				return nil
			}, nil
		},
	}

	svc := NewAnalysisService(db, nil, mockResolver)
	ctx := context.Background()

	// Expect migration audit insert
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO migration_audits (project_id, migration_name, status, findings_json) VALUES ($1, $2, $3, $4) RETURNING id")).
		WithArgs("proj-1", "001_drop_col", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("audit-1"))

	// Expect index recommendation upsert
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO index_recommendations (project_id, table_name, column_names, reason, benefit_score, applied) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (project_id, table_name, column_names) DO UPDATE SET reason = EXCLUDED.reason, benefit_score = EXCLUDED.benefit_score, applied = EXCLUDED.applied, updated_at = NOW() RETURNING id")).
		WithArgs("proj-1", "users", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), false).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("rec-1"))

	// Act
	report, err := svc.RunProjectAnalysis(ctx, ProjectAnalysisRequest{
		ProjectID:     "proj-1",
		MigrationName: "001_drop_col",
		MigrationSQL:  "ALTER TABLE users DROP COLUMN email",
		RoleOpts:      analyzer.AnalyzeOptions{},
		ValOpts:       analyzer.ValidationOptions{CheckBreaking: true},
		IdxOpts:       analyzer.IndexAnalysisOptions{MinQueryExecutions: 100, MinTableSize: 1},
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if report == nil || report.Migration == nil || report.Index == nil || report.Role == nil {
		t.Fatal("expected non-nil report parts")
	}
	if report.Migration.Status != "fail" {
		t.Errorf("expected migration status 'fail', got '%s'", report.Migration.Status)
	}
	if len(report.Index.Recommendations) == 0 {
		t.Error("expected at least one index recommendation")
	}
	if !closerCalled {
		t.Error("expected closer to be called")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestAnalysisService_RunProjectAnalysis_ResolveFails(t *testing.T) {
	// Arrange
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mockResolver := &MockProjectResolver{
		ResolveInspectorFunc: func(ctx context.Context, projectID string) (analyzer.DBInspector, func() error, error) {
			return nil, nil, errors.New("vault is sealed")
		},
	}

	svc := NewAnalysisService(db, nil, mockResolver)
	ctx := context.Background()

	// Act
	_, err = svc.RunProjectAnalysis(ctx, ProjectAnalysisRequest{
		ProjectID:     "proj-1",
		MigrationName: "001_test",
		MigrationSQL:  "SELECT 1",
	})

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expectedErr := "failed to resolve project database: vault is sealed"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestAnalysisService_RunProjectAnalysis_CloserCalled(t *testing.T) {
	// Arrange
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	closerCalled := false
	// Mock resolver that returns errorInspector
	mockResolver := &MockProjectResolver{
		ResolveInspectorFunc: func(ctx context.Context, projectID string) (analyzer.DBInspector, func() error, error) {
			return &errorInspector{}, func() error {
				closerCalled = true
				return nil
			}, nil
		},
	}

	svc := NewAnalysisService(db, nil, mockResolver)
	ctx := context.Background()

	// Act
	_, err = svc.RunProjectAnalysis(ctx, ProjectAnalysisRequest{
		ProjectID:     "proj-1",
		MigrationName: "001_test",
		MigrationSQL:  "SELECT 1",
	})

	// Assert
	if err == nil {
		t.Fatal("expected error from analyzer")
	}
	if !closerCalled {
		t.Error("expected closer to be called even on analyzer error")
	}
}

func TestAnalysisService_RunProjectAnalysis_EmptyProjectID(t *testing.T) {
	// Arrange
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	svc := NewAnalysisService(db, nil, nil)
	ctx := context.Background()

	// Act
	_, err = svc.RunProjectAnalysis(ctx, ProjectAnalysisRequest{
		ProjectID:     "",
		MigrationName: "001_test",
		MigrationSQL:  "SELECT 1",
	})

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	expectedErr := "projectID is required"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}
