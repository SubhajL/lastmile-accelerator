package analyzer

import (
	"context"
	"fmt"
	"testing"
)

// mockInspectorIndex is a focused mock for index advisor tests
type mockInspectorIndex struct {
	queryStats []QueryStat
	indexes   map[string][]IndexInfo
	tables    []TableInfo
}

type erroringInspector struct{ mockInspectorIndex }

func (e *erroringInspector) GetQueryStats(ctx context.Context, minExecutions int) ([]QueryStat, error) {
	return nil, fmt.Errorf("pg_stat_statements unavailable")
}

func (m *mockInspectorIndex) GetTables(ctx context.Context, schema string) ([]TableInfo, error) {
	return m.tables, nil
}
func (m *mockInspectorIndex) GetTableColumns(ctx context.Context, table string) ([]ColumnInfo, error) {
	return nil, nil
}
func (m *mockInspectorIndex) GetIndexes(ctx context.Context, table string) ([]IndexInfo, error) {
	return m.indexes[table], nil
}
func (m *mockInspectorIndex) GetRoles(ctx context.Context) ([]RoleInfo, error) { return nil, nil }
func (m *mockInspectorIndex) GetQueryStats(ctx context.Context, minExecutions int) ([]QueryStat, error) {
	// Return only stats with >= minExecutions
	var out []QueryStat
	for _, s := range m.queryStats {
		if int(s.Calls) >= minExecutions {
			out = append(out, s)
		}
	}
	return out, nil
}
func (m *mockInspectorIndex) AnalyzeMigration(ctx context.Context, sql string) (*MigrationAnalysis, error) {
	return &MigrationAnalysis{}, nil
}

func TestIndexAdvisor_AnalyzeIndexes_FrequentSeqScan(t *testing.T) {
	// Arrange: frequent WHERE users.email = ?
	mock := &mockInspectorIndex{
		queryStats: []QueryStat{
			{Query: "SELECT * FROM users WHERE email = $1", Calls: 500, TotalTime: 10000, MeanTime: 20},
		},
		indexes: map[string][]IndexInfo{
			"users": {},
		},
		tables: []TableInfo{{Schema: "public", Name: "users", SizeBytes: 50 * 1024 * 1024}}, // 50MB
	}
	advisor := NewIndexAdvisor(mock)
	ctx := context.Background()

	// Act
	recs, err := advisor.AnalyzeIndexes(ctx, IndexAnalysisOptions{MinQueryExecutions: 100, MinTableSize: 1})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(recs.Recommendations) == 0 {
		t.Fatalf("expected at least one recommendation, got none")
	}
	rec := recs.Recommendations[0]
	if rec.TableName != "users" {
		t.Errorf("expected table 'users', got '%s'", rec.TableName)
	}
	if len(rec.Columns) != 1 || rec.Columns[0] != "email" {
		t.Errorf("expected index on (email), got %v", rec.Columns)
	}
	if rec.BenefitScore <= 0 {
		t.Error("expected positive benefit score")
	}
}

func TestIndexAdvisor_AnalyzeIndexes_DuplicateIndex_NotRecommended(t *testing.T) {
	// Arrange: existing index on orders(user_id)
	mock := &mockInspectorIndex{
		queryStats: []QueryStat{
			{Query: "SELECT * FROM orders WHERE user_id = $1", Calls: 200, TotalTime: 5000, MeanTime: 25},
		},
		indexes: map[string][]IndexInfo{
			"orders": {{Name: "idx_orders_user_id", TableName: "orders", Columns: []string{"user_id"}, IsUnique: false}},
		},
		tables: []TableInfo{{Schema: "public", Name: "orders", SizeBytes: 30 * 1024 * 1024}},
	}
	advisor := NewIndexAdvisor(mock)
	ctx := context.Background()

	// Act
	recs, err := advisor.AnalyzeIndexes(ctx, IndexAnalysisOptions{MinQueryExecutions: 100, MinTableSize: 1, CheckDuplicates: true})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(recs.Recommendations) != 0 {
		t.Fatalf("expected no recommendations due to existing index, got %d", len(recs.Recommendations))
	}
}

func TestIndexAdvisor_Scoring_IncreasesWithCalls(t *testing.T) {
	// Arrange
	mock := &mockInspectorIndex{
		queryStats: []QueryStat{
			{Query: "select * from users where last_login > $1", Calls: 50, TotalTime: 500, MeanTime: 10},
			{Query: "select * from users where last_login > $1", Calls: 1000, TotalTime: 30000, MeanTime: 30},
		},
		indexes: map[string][]IndexInfo{"users": {}},
		tables:  []TableInfo{{Schema: "public", Name: "users", SizeBytes: 100 * 1024 * 1024}},
	}
	advisor := NewIndexAdvisor(mock)
	ctx := context.Background()

	// Act
	recs, err := advisor.AnalyzeIndexes(ctx, IndexAnalysisOptions{MinQueryExecutions: 10, MinTableSize: 1})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(recs.Recommendations) == 0 {
		t.Fatalf("expected recommendations, got none")
	}
	// Expect higher score from higher call count
	if recs.Recommendations[1].BenefitScore <= recs.Recommendations[0].BenefitScore {
		t.Error("expected benefit score to increase with calls")
	}
}

func TestIndexAdvisor_AnalyzeIndexes_MultiColumnPredicate_PicksFirstColumn(t *testing.T) {
	mock := &mockInspectorIndex{
		queryStats: []QueryStat{{Query: "select * from orders where user_id = $1 and status = $2", Calls: 150}},
		indexes:   map[string][]IndexInfo{"orders": {}},
		tables:    []TableInfo{{Schema: "public", Name: "orders", SizeBytes: 10 * 1024 * 1024}},
	}
	advisor := NewIndexAdvisor(mock)
	ctx := context.Background()

	recs, err := advisor.AnalyzeIndexes(ctx, IndexAnalysisOptions{MinQueryExecutions: 100, MinTableSize: 1})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(recs.Recommendations) == 0 || recs.Recommendations[0].Columns[0] != "user_id" {
		t.Fatalf("expected first predicate column user_id, got %+v", recs.Recommendations)
	}
}

func TestIndexAdvisor_AnalyzeIndexes_Like_And_Ilike(t *testing.T) {
	mock := &mockInspectorIndex{
		queryStats: []QueryStat{
			{Query: "select * from users where email like $1", Calls: 120},
			{Query: "select * from users where name ilike $1", Calls: 120},
		},
		indexes: map[string][]IndexInfo{"users": {}},
		tables:  []TableInfo{{Schema: "public", Name: "users", SizeBytes: 20 * 1024 * 1024}},
	}
	advisor := NewIndexAdvisor(mock)
	ctx := context.Background()
	recs, err := advisor.AnalyzeIndexes(ctx, IndexAnalysisOptions{MinQueryExecutions: 100, MinTableSize: 1})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(recs.Recommendations) < 2 {
		t.Fatalf("expected at least 2 recs, got %d", len(recs.Recommendations))
	}
}

func TestIndexAdvisor_AnalyzeIndexes_InList_Filter(t *testing.T) {
	mock := &mockInspectorIndex{
		queryStats: []QueryStat{{Query: "select * from log where level in ($1,$2,$3)", Calls: 300}},
		indexes:   map[string][]IndexInfo{"log": {}},
		tables:    []TableInfo{{Schema: "public", Name: "log", SizeBytes: 80 * 1024 * 1024}},
	}
	advisor := NewIndexAdvisor(mock)
	ctx := context.Background()
	recs, err := advisor.AnalyzeIndexes(ctx, IndexAnalysisOptions{MinQueryExecutions: 100, MinTableSize: 1})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(recs.Recommendations) == 0 || recs.Recommendations[0].Columns[0] != "level" {
		t.Fatalf("expected level column recommendation")
	}
}

func TestIndexAdvisor_AnalyzeIndexes_TableSizeThreshold_SkipsSmallTables(t *testing.T) {
	mock := &mockInspectorIndex{
		queryStats: []QueryStat{{Query: "select * from small where k = $1", Calls: 1000}},
		indexes:   map[string][]IndexInfo{"small": {}},
		tables:    []TableInfo{{Schema: "public", Name: "small", SizeBytes: 1024}},
	}
	advisor := NewIndexAdvisor(mock)
	ctx := context.Background()
	recs, err := advisor.AnalyzeIndexes(ctx, IndexAnalysisOptions{MinQueryExecutions: 10, MinTableSize: 10 * 1024 * 1024})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(recs.Recommendations) != 0 {
		t.Fatalf("expected no recommendations due to size threshold")
	}
}

func TestIndexAdvisor_GracefulWhenQueryStatsUnavailable_ReturnsNoRecs(t *testing.T) {
	// Arrange erroring inspector simulates missing pg_stat_statements
	base := mockInspectorIndex{
		queryStats: nil,
		indexes:   map[string][]IndexInfo{"users": {}},
		tables:    []TableInfo{{Schema: "public", Name: "users", SizeBytes: 50 * 1024 * 1024}},
	}
	e := &erroringInspector{mockInspectorIndex: base}
	advisor := NewIndexAdvisor(e)
	ctx := context.Background()
	// Act
	recs, err := advisor.AnalyzeIndexes(ctx, IndexAnalysisOptions{MinQueryExecutions: 100, MinTableSize: 1})
	// Assert
	if err != nil {
		t.Fatalf("expected graceful nil error, got %v", err)
	}
	if len(recs.Recommendations) != 0 {
		t.Fatalf("expected 0 recommendations when stats unavailable")
	}
}
