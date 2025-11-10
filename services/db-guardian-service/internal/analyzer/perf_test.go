package analyzer

import (
	"context"
	"testing"
	"time"
)

func TestPerformance_ValidateMigration_Timed(t *testing.T) {
	mock := &mockInspector{}
	guard := NewMigrationGuard(mock)
	ctx := context.Background()
	sql := "ALTER TABLE users ADD COLUMN c1 TEXT; ALTER TABLE users ADD COLUMN c2 TEXT; ALTER TABLE users DROP COLUMN c1"
	opts := ValidationOptions{CheckBreaking: true, CheckPerformance: false}

	iters := 50
	start := time.Now()
	for i := 0; i < iters; i++ {
		if _, err := guard.ValidateMigration(ctx, sql, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	dur := time.Since(start)
	avg := dur / time.Duration(iters)
	// Non-flaky generous budget: avg under 50ms
	if avg > 50*time.Millisecond {
		t.Fatalf("ValidateMigration average too high: %v", avg)
	}
}

func TestPerformance_AnalyzeIndexes_Timed(t *testing.T) {
	mock := &mockInspectorIndex{
		queryStats: []QueryStat{
			{Query: "select * from users where email = $1", Calls: 1000},
			{Query: "select * from users where last_login > $1", Calls: 800},
		},
		indexes: map[string][]IndexInfo{"users": {}},
		tables:  []TableInfo{{Schema: "public", Name: "users", SizeBytes: 200 * 1024 * 1024}},
	}
	advisor := NewIndexAdvisor(mock)
	ctx := context.Background()
	opts := IndexAnalysisOptions{MinQueryExecutions: 10, MinTableSize: 1}

	iters := 50
	start := time.Now()
	for i := 0; i < iters; i++ {
		if _, err := advisor.AnalyzeIndexes(ctx, opts); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	dur := time.Since(start)
	avg := dur / time.Duration(iters)
	// Generous budget: avg under 50ms
	if avg > 50*time.Millisecond {
		t.Fatalf("AnalyzeIndexes average too high: %v", avg)
	}
}
