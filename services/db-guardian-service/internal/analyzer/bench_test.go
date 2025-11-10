package analyzer

import (
	"context"
	"testing"
)

// Bench: ValidateMigration on typical SQL
func BenchmarkValidateMigration_Simple(b *testing.B) {
	mock := &mockInspector{}
	guard := NewMigrationGuard(mock)
	ctx := context.Background()
	sql := "ALTER TABLE users ADD COLUMN phone VARCHAR(20); ALTER TABLE users DROP COLUMN legacy"
	opts := ValidationOptions{CheckBreaking: true, CheckPerformance: false}
	for i := 0; i < b.N; i++ {
		_, _ = guard.ValidateMigration(ctx, sql, opts)
	}
}

// Bench: AnalyzeIndexes on frequent filter
func BenchmarkAnalyzeIndexes_Simple(b *testing.B) {
	mock := &mockInspectorIndex{
		queryStats: []QueryStat{{Query: "select * from users where email = $1", Calls: 1000}},
		indexes:   map[string][]IndexInfo{"users": {}},
		tables:    []TableInfo{{Schema: "public", Name: "users", SizeBytes: 100 * 1024 * 1024}},
	}
	advisor := NewIndexAdvisor(mock)
	ctx := context.Background()
	opts := IndexAnalysisOptions{MinQueryExecutions: 10, MinTableSize: 1}
	for i := 0; i < b.N; i++ {
		_, _ = advisor.AnalyzeIndexes(ctx, opts)
	}
}
