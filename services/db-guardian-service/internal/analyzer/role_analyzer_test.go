package analyzer

import (
	"context"
	"strings"
	"testing"
	"time"
)

// mockInspector is a test double for DBInspector
type mockInspector struct {
	roles      []RoleInfo
	tables     []TableInfo
	rolesErr   error
	tablesErr  error
}

func (m *mockInspector) GetTables(ctx context.Context, schema string) ([]TableInfo, error) {
	return m.tables, m.tablesErr
}

func (m *mockInspector) GetTableColumns(ctx context.Context, table string) ([]ColumnInfo, error) {
	return nil, nil
}

func (m *mockInspector) GetIndexes(ctx context.Context, table string) ([]IndexInfo, error) {
	return nil, nil
}

func (m *mockInspector) GetRoles(ctx context.Context) ([]RoleInfo, error) {
	return m.roles, m.rolesErr
}

func (m *mockInspector) GetQueryStats(ctx context.Context, minExecutions int) ([]QueryStat, error) {
	return nil, nil
}

func (m *mockInspector) AnalyzeMigration(ctx context.Context, sql string) (*MigrationAnalysis, error) {
	// Simple SQL parsing for testing
	analysis := &MigrationAnalysis{
		SQL:        sql,
		Operations: []SQLOperation{},
	}

	// Detect breaking changes from SQL
	sqlUpper := strings.ToUpper(sql)
	if strings.Contains(sqlUpper, "DROP COLUMN") || strings.Contains(sqlUpper, "DROP TABLE") {
		analysis.HasBreaking = true
		op := SQLOperation{
			Type:       "ALTER",
			ObjectType: "COLUMN",
			Details:    map[string]string{"is_breaking": "true"},
		}
		analysis.Operations = append(analysis.Operations, op)
	}

	if strings.Contains(sqlUpper, "CREATE TABLE") {
		op := SQLOperation{
			Type:       "CREATE",
			ObjectType: "TABLE",
			Details:    map[string]string{"is_breaking": "false"},
		}
		// Extract table name
		if idx := strings.Index(sqlUpper, "TABLE "); idx != -1 {
			rest := sqlUpper[idx+6:]
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				op.ObjectName = strings.ToLower(parts[0])
			}
		}
		analysis.Operations = append(analysis.Operations, op)
	}

	return analysis, nil
}

func TestRoleAnalyzer_Analyze_SuperuserDetected(t *testing.T) {
	// Arrange
	mock := &mockInspector{
		roles: []RoleInfo{
			{
				Name:        "app_user",
				IsSuperuser: true,
				CanLogin:    true,
				Permissions: []Permission{
					{ObjectType: "database", ObjectName: "*", Privilege: "SUPERUSER"},
				},
			},
		},
	}

	analyzer := NewRoleAnalyzer(mock)
	ctx := context.Background()
	opts := AnalyzeOptions{
		IncludeSystemRoles: false,
		CheckPublicSchema:  true,
	}

	// Act
	result, err := analyzer.Analyze(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Findings) == 0 {
		t.Fatal("expected findings for superuser, got none")
	}
	
	hasCritical := false
	for _, f := range result.Findings {
		if f.Severity == SeverityCritical && f.RoleName == "app_user" {
			hasCritical = true
			break
		}
	}
	if !hasCritical {
		t.Error("expected critical finding for superuser role")
	}
}

func TestRoleAnalyzer_Analyze_NoRoles_ReturnsEmptyFindings(t *testing.T) {
	// Arrange
	mock := &mockInspector{
		roles: []RoleInfo{},
	}

	analyzer := NewRoleAnalyzer(mock)
	ctx := context.Background()
	opts := AnalyzeOptions{}

	// Act
	result, err := analyzer.Analyze(ctx, opts)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("expected no findings for no roles, got %d", len(result.Findings))
	}
	if result.RolesAnalyzed != 0 {
		t.Errorf("expected 0 roles analyzed, got %d", result.RolesAnalyzed)
	}
}

func TestRoleAnalyzer_Analyze_CreateRolePermission_WarnsHighPrivilege(t *testing.T) {
	// Arrange
	mock := &mockInspector{
		roles: []RoleInfo{
			{
				Name:        "admin_user",
				IsSuperuser: false,
				CanLogin:    true,
				Permissions: []Permission{
					{ObjectType: "database", ObjectName: "*", Privilege: "CREATEROLE"},
				},
			},
		},
	}

	analyzer := NewRoleAnalyzer(mock)
	ctx := context.Background()

	// Act
	result, err := analyzer.Analyze(ctx, AnalyzeOptions{})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	
	hasWarn := false
	for _, f := range result.Findings {
		if f.Severity == SeverityWarn && f.RoleName == "admin_user" {
			hasWarn = true
			break
		}
	}
	if !hasWarn {
		t.Error("expected warning finding for CREATEROLE privilege")
	}
}

func TestRoleAnalyzer_Analyze_Createdb_Info(t *testing.T) {
	mock := &mockInspector{roles: []RoleInfo{{Name: "builder", CanLogin: true, Permissions: []Permission{{Privilege: "CREATEDB", ObjectType: "database", ObjectName: "*"}}}}}
	an := NewRoleAnalyzer(mock)
	res, err := an.Analyze(context.Background(), AnalyzeOptions{})
	if err != nil { t.Fatalf("unexpected: %v", err) }
	found := false
	for _, f := range res.Findings {
		if f.Severity == SeverityInfo && f.RoleName == "builder" {
			found = true
		}
	}
	if !found { t.Fatalf("expected CREATEDB info finding") }
}

func TestRoleAnalyzer_GeneratePolicy_ProducesYAML(t *testing.T) {
	// Arrange
	mock := &mockInspector{
		roles: []RoleInfo{
			{
				Name:        "readonly_user",
				IsSuperuser: false,
				CanLogin:    true,
				Permissions: []Permission{
					{Schema: "public", ObjectType: "table", ObjectName: "users", Privilege: "SELECT"},
				},
			},
		},
	}

	analyzer := NewRoleAnalyzer(mock)
	ctx := context.Background()

	// Act
	result, err := analyzer.Analyze(ctx, AnalyzeOptions{})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.PolicyYAML == "" {
		t.Error("expected non-empty policy YAML")
	}
	if result.AnalyzedAt.IsZero() {
		t.Error("expected AnalyzedAt timestamp to be set")
	}
}

func TestRoleAnalyzer_Analyze_ResultTimestamp_IsSet(t *testing.T) {
	// Arrange
	mock := &mockInspector{
		roles: []RoleInfo{},
	}

	analyzer := NewRoleAnalyzer(mock)
	ctx := context.Background()
	before := time.Now()

	// Act
	result, _ := analyzer.Analyze(ctx, AnalyzeOptions{})

	// Assert
	after := time.Now()
	if result.AnalyzedAt.Before(before) || result.AnalyzedAt.After(after) {
		t.Error("expected AnalyzedAt to be set to current time")
	}
}
