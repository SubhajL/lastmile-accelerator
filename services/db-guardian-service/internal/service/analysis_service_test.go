package service

import (
	"context"
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

func TestAnalysisService_RunFullAnalysis_Persists(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	insp := &fakeInspector{}
	rAnalyzer := analyzer.NewRoleAnalyzer(insp)
	mGuard := analyzer.NewMigrationGuard(insp)
	iAdvisor := analyzer.NewIndexAdvisor(insp)

	svc := NewAnalysisService(db, nil, rAnalyzer, mGuard, iAdvisor)
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
	report, err := svc.RunFullAnalysis(ctx, "proj-1", "001_drop_col", "ALTER TABLE users DROP COLUMN email", analyzer.AnalyzeOptions{}, analyzer.ValidationOptions{CheckBreaking: true}, analyzer.IndexAnalysisOptions{MinQueryExecutions: 100, MinTableSize: 1})

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
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
