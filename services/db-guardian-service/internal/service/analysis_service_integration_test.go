package service

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nats-io/nats.go"
)

type capturePublisher struct{ subjects []string }

// fakeInspectorFull produces one breaking migration and one index recommendation
type fakeInspectorFull struct{}

func (f *fakeInspectorFull) GetTables(ctx context.Context, schema string) ([]analyzer.TableInfo, error) {
	return []analyzer.TableInfo{{Schema: "public", Name: "users", SizeBytes: 64 * 1024 * 1024}}, nil
}
func (f *fakeInspectorFull) GetTableColumns(ctx context.Context, table string) ([]analyzer.ColumnInfo, error) {
	return nil, nil
}
func (f *fakeInspectorFull) GetIndexes(ctx context.Context, table string) ([]analyzer.IndexInfo, error) {
	return []analyzer.IndexInfo{}, nil
}
func (f *fakeInspectorFull) GetRoles(ctx context.Context) ([]analyzer.RoleInfo, error) {
	return []analyzer.RoleInfo{{Name: "app", IsSuperuser: true, CanLogin: true}}, nil
}
func (f *fakeInspectorFull) GetQueryStats(ctx context.Context, minExecutions int) ([]analyzer.QueryStat, error) {
	return []analyzer.QueryStat{{Query: "select * from users where email = $1", Calls: 500}}, nil
}
func (f *fakeInspectorFull) AnalyzeMigration(ctx context.Context, sql string) (*analyzer.MigrationAnalysis, error) {
	return &analyzer.MigrationAnalysis{SQL: sql, HasBreaking: true, Operations: []analyzer.SQLOperation{{Type: "ALTER", ObjectType: "COLUMN", ObjectName: "users", Details: map[string]string{"operation": "DROP", "is_breaking": "true"}}}}, nil
}

func TestAnalysisService_RunFullAnalysis_PersistsAndPublishes(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock: %v", err) }
	defer db.Close()

	insp := &fakeInspectorFull{}
	rAnalyzer := analyzer.NewRoleAnalyzer(insp)
	mGuard := analyzer.NewMigrationGuard(insp)
	iAdvisor := analyzer.NewIndexAdvisor(insp)

dummyNats := &nats.Conn{}
svc := NewAnalysisService(db, dummyNats, rAnalyzer, mGuard, iAdvisor)
cap := &capturePublisher{}
	// adapt to EventPublisher signature
svc.SetPublisher(func(ctx context.Context, _ *nats.Conn, subject string, data []byte) error {
		cap.subjects = append(cap.subjects, subject)
		// validate payload is JSON
		var tmp any
		_ = json.Unmarshal(data, &tmp)
		return nil
	})

	ctx := context.Background()

	// Expect migration audit insert
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO migration_audits (project_id, migration_name, status, findings_json) VALUES ($1, $2, $3, $4) RETURNING id")).
		WithArgs("proj-2", "003", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("audit-9"))
	// Expect index recommendation upsert
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO index_recommendations (project_id, table_name, column_names, reason, benefit_score, applied) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (project_id, table_name, column_names) DO UPDATE SET reason = EXCLUDED.reason, benefit_score = EXCLUDED.benefit_score, applied = EXCLUDED.applied, updated_at = NOW() RETURNING id")).
		WithArgs("proj-2", "users", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), false).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("rec-9"))

	// Act
	report, err := svc.RunFullAnalysis(ctx, "proj-2", "003", "ALTER TABLE users DROP COLUMN email", analyzer.AnalyzeOptions{}, analyzer.ValidationOptions{CheckBreaking: true}, analyzer.IndexAnalysisOptions{MinQueryExecutions: 100, MinTableSize: 1})

	// Assert
	if err != nil { t.Fatalf("unexpected error: %v", err) }
	if report == nil || report.Migration == nil || report.Index == nil || report.Role == nil {
		t.Fatalf("expected full report")
	}
	if len(cap.subjects) != 3 {
		t.Fatalf("expected 3 events, got %d", len(cap.subjects))
	}
	want := map[string]bool{"db.analysis.completed":true, "db.migration.validated":true, "db.index.recommended":true}
	for _, s := range cap.subjects { if !want[s] { t.Fatalf("unexpected subject %s", s) } }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("sql expectations: %v", err) }
}
