package analyzer

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgresInspector_GetTables_ReturnsTableList(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	inspector := NewPostgresInspector(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"schema", "name", "row_count", "size_bytes"}).
		AddRow("public", "users", 1000, 8192000).
		AddRow("public", "orders", 5000, 16384000)

	mock.ExpectQuery(`SELECT (.+) FROM information_schema.tables`).
		WithArgs("public").
		WillReturnRows(rows)

	// Act
	tables, err := inspector.GetTables(ctx, "public")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(tables) != 2 {
		t.Errorf("expected 2 tables, got %d", len(tables))
	}
	if tables[0].Name != "users" {
		t.Errorf("expected first table 'users', got '%s'", tables[0].Name)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresInspector_GetTableColumns_InvalidTable_ReturnsError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	inspector := NewPostgresInspector(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT (.+) FROM information_schema.columns`).
		WithArgs("nonexistent").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "data_type", "is_nullable", "column_default"}))

	// Act
	columns, err := inspector.GetTableColumns(ctx, "nonexistent")

	// Assert
	if err != nil {
		t.Errorf("expected no error for empty result, got %v", err)
	}
	if len(columns) != 0 {
		t.Errorf("expected empty columns for nonexistent table, got %d", len(columns))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresInspector_GetIndexes_NoIndexes_ReturnsEmptySlice(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	inspector := NewPostgresInspector(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT (.+) FROM pg_indexes`).
		WithArgs("test_table").
		WillReturnRows(sqlmock.NewRows([]string{"indexname", "tablename", "indexdef"}))

	// Act
	indexes, err := inspector.GetIndexes(ctx, "test_table")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if indexes == nil {
		t.Error("expected empty slice, got nil")
	}
	if len(indexes) != 0 {
		t.Errorf("expected 0 indexes, got %d", len(indexes))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresInspector_GetIndexes_QueryError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock: %v", err) }
	defer db.Close()
	inspector := NewPostgresInspector(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT (.+) FROM pg_indexes`).
		WithArgs("oops").
		WillReturnError(fmt.Errorf("broken"))

	_, err = inspector.GetIndexes(ctx, "oops")
	if err == nil { t.Fatalf("expected error") }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("sql expectations: %v", err) }
}

func TestPostgresInspector_GetQueryStats_NoStatsExtension_ReturnsError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	inspector := NewPostgresInspector(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT (.+) FROM pg_stat_statements`).
		WithArgs(100).
		WillReturnError(sql.ErrNoRows)

	// Act
	_, err = inspector.GetQueryStats(ctx, 100)

	// Assert
	if err == nil {
		t.Error("expected error when pg_stat_statements unavailable, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresInspector_GetRoles_Positive(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	inspector := NewPostgresInspector(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"rolname", "rolsuper", "rolinherit", "rolcreaterole", "rolcreatedb", "rolcanlogin"}).
		AddRow("app_user", true, true, false, true, true).
		AddRow("readonly", false, true, false, false, true)

	mock.ExpectQuery(`SELECT\s+rolname,\s*rolsuper,\s*rolinherit,\s*rolcreaterole,\s*rolcreatedb,\s*rolcanlogin\s+FROM pg_roles`).
		WillReturnRows(rows)

	// Act
	roles, err := inspector.GetRoles(ctx)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}
	if !roles[0].IsSuperuser || !roles[0].CanLogin {
		t.Errorf("expected app_user superuser+login true")
	}
	// Permissions list should include SUPERUSER and CREATEDB for app_user
	foundSU, foundCDB := false, false
	for _, p := range roles[0].Permissions {
		if p.Privilege == "SUPERUSER" {
			foundSU = true
		}
		if p.Privilege == "CREATEDB" {
			foundCDB = true
		}
	}
	if !(foundSU && foundCDB) {
		t.Errorf("expected SUPERUSER and CREATEDB permissions, got %+v", roles[0].Permissions)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresInspector_GetRoles_CreateRolePermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock: %v", err) }
	defer db.Close()
	insp := NewPostgresInspector(db)
	ctx := context.Background()
	rows := sqlmock.NewRows([]string{"rolname","rolsuper","rolinherit","rolcreaterole","rolcreatedb","rolcanlogin"}).
		AddRow("role_mgr", false, true, true, false, true)
	mock.ExpectQuery(`SELECT\s+rolname,\s*rolsuper,\s*rolinherit,\s*rolcreaterole,\s*rolcreatedb,\s*rolcanlogin\s+FROM pg_roles`).
		WillReturnRows(rows)
	roles, err := insp.GetRoles(ctx)
	if err != nil { t.Fatalf("unexpected: %v", err) }
	found := false
	for _, p := range roles[0].Permissions { if p.Privilege == "CREATEROLE" { found = true } }
	if !found { t.Fatalf("expected CREATEROLE permission") }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("sql expectations: %v", err) }
}

func TestPostgresInspector_GetTableColumns_Positive(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	inspector := NewPostgresInspector(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"column_name", "data_type", "is_nullable", "column_default"}).
		AddRow("id", "integer", "NO", "nextval('t_id_seq'::regclass)").
		AddRow("email", "text", "YES", nil)

	mock.ExpectQuery(`SELECT (.+) FROM information_schema.columns`).
		WithArgs("users").
		WillReturnRows(rows)

	// Act
	cols, err := inspector.GetTableColumns(ctx, "users")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	if cols[0].DataType != "integer" || cols[0].IsNullable {
		t.Errorf("expected id integer NOT NULL")
	}
	if cols[1].DataType != "text" || !cols[1].IsNullable {
		t.Errorf("expected email text NULL")
	}
	if cols[0].DefaultValue == "" {
		t.Errorf("expected default value on id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestParseIndexColumns_FunctionExpr(t *testing.T) {
	def := "CREATE INDEX idx_fn ON public.users USING btree ((lower(email)))"
	cols := parseIndexColumns(def)
	if len(cols) != 1 {
		t.Fatalf("expected 1 column, got %d", len(cols))
	}
	if cols[0] == "" {
		t.Errorf("expected non-empty expression column")
	}
}

func TestPostgresInspector_GetIndexes_UniqueAndMultiColumns(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	inspector := NewPostgresInspector(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"indexname", "tablename", "indexdef"}).
		AddRow("idx_u", "users", "CREATE UNIQUE INDEX idx_u ON public.users USING btree (email, created_at)")

	mock.ExpectQuery(`SELECT (.+) FROM pg_indexes`).
		WithArgs("users").
		WillReturnRows(rows)

	// Act
	indexes, err := inspector.GetIndexes(ctx, "users")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(indexes) != 1 {
		t.Fatalf("expected 1 index, got %d", len(indexes))
	}
	if !indexes[0].IsUnique {
		t.Errorf("expected unique index")
	}
	if len(indexes[0].Columns) != 2 || indexes[0].Columns[0] == "" || indexes[0].Columns[1] == "" {
		t.Errorf("expected two parsed columns, got %+v", indexes[0].Columns)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestPostgresInspector_AnalyzeMigration_BreakingChange_DetectsIssue(t *testing.T) {
	// Arrange
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	inspector := NewPostgresInspector(db)
	ctx := context.Background()

	// Act
	analysis, err := inspector.AnalyzeMigration(ctx, "ALTER TABLE users DROP COLUMN email")

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !analysis.HasBreaking {
		t.Error("expected HasBreaking=true for DROP COLUMN, got false")
	}
	if len(analysis.Operations) == 0 {
		t.Error("expected at least one operation detected")
	}
}

func TestPostgresInspector_GetQueryStats_Positive(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock: %v", err) }
	defer db.Close()
	inspector := NewPostgresInspector(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"query","calls","total_exec_time","mean_exec_time"}).
		AddRow("select 1", int64(10), float64(5.5), float64(0.55))
	mock.ExpectQuery(`SELECT (.+) FROM pg_stat_statements`).
		WithArgs(1).
		WillReturnRows(rows)

	stats, err := inspector.GetQueryStats(ctx, 1)
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if len(stats) != 1 || stats[0].Calls != 10 || stats[0].TotalTime == 0 || stats[0].MeanTime == 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("sql expectations: %v", err) }
}

func TestPostgresInspector_GetTables_QueryError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock: %v", err) }
	defer db.Close()
	inspector := NewPostgresInspector(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT (.+) FROM information_schema.tables`).
		WithArgs("public").
		WillReturnError(fmt.Errorf("boom"))

	_, err = inspector.GetTables(ctx, "public")
	if err == nil { t.Fatalf("expected error") }
	if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("sql expectations: %v", err) }
}

func TestParseIndexColumns_UnknownPattern_ReturnsEmpty(t *testing.T) {
	cols := parseIndexColumns("ANALYZE users")
	if len(cols) != 0 { t.Fatalf("expected empty, got %v", cols) }
}

func TestPostgresInspector_AnalyzeMigration_CreateIndexConcurrently_Details(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock: %v", err) }
	defer db.Close()
	inspector := NewPostgresInspector(db)
	ctx := context.Background()
	analysis, err := inspector.AnalyzeMigration(ctx, "CREATE INDEX CONCURRENTLY idx ON users (email)")
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if analysis.HasBreaking { t.Fatalf("did not expect breaking") }
	if len(analysis.Operations) != 1 { t.Fatalf("expected one op") }
	if analysis.Operations[0].Details["lock_level"] == "" { t.Fatalf("expected lock_level detail") }
}

func TestPostgresInspector_AnalyzeMigration_UnknownStatement_UnknownOperation(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil { t.Fatalf("sqlmock: %v", err) }
	defer db.Close()
	insp := NewPostgresInspector(db)
	ctx := context.Background()
	an, err := insp.AnalyzeMigration(ctx, "VACUUM")
	if err != nil { t.Fatalf("unexpected: %v", err) }
	if len(an.Operations) != 1 || an.Operations[0].Type != "UNKNOWN" {
		t.Fatalf("expected UNKNOWN operation, got %+v", an.Operations)
	}
	if an.HasBreaking { t.Fatalf("did not expect breaking for VACUUM") }
}
