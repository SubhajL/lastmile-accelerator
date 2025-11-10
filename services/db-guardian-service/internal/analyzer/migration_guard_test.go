package analyzer

import (
	"context"
	"strings"
	"testing"
)

func TestMigrationGuard_ValidateMigration_DropColumn_Fails(t *testing.T) {
	// Arrange
	mock := &mockInspector{}
	guard := NewMigrationGuard(mock)
	ctx := context.Background()

	sql := "ALTER TABLE users DROP COLUMN email"
	opts := ValidationOptions{
		CheckBreaking:    true,
		CheckPerformance: false,
	}

	// Act
	result, err := guard.ValidateMigration(ctx, sql, opts)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Status != "fail" {
		t.Errorf("expected status 'fail' for DROP COLUMN, got '%s'", result.Status)
	}
	if len(result.Findings) == 0 {
		t.Error("expected findings for breaking change")
	}
}

func TestMigrationGuard_ValidateMigration_AddColumn_Passes(t *testing.T) {
	// Arrange
	mock := &mockInspector{}
	guard := NewMigrationGuard(mock)
	ctx := context.Background()

	sql := "ALTER TABLE users ADD COLUMN phone VARCHAR(20)"
	opts := ValidationOptions{
		CheckBreaking:    true,
		CheckPerformance: false,
	}

	// Act
	result, err := guard.ValidateMigration(ctx, sql, opts)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Status == "fail" {
		t.Errorf("expected status 'pass' or 'warn' for ADD COLUMN, got '%s'", result.Status)
	}
}

func TestMigrationGuard_ValidateMigration_LargeTable_WarnsLock(t *testing.T) {
	// Arrange
	mock := &mockInspector{
		tables: []TableInfo{
			{Name: "orders", SizeBytes: 100 * 1024 * 1024}, // 100 MB
		},
	}
	guard := NewMigrationGuard(mock)
	ctx := context.Background()

	sql := "ALTER TABLE orders ADD COLUMN status VARCHAR(20)"
	opts := ValidationOptions{
		CheckBreaking:    false,
		CheckPerformance: true,
		MaxTableSize:     10 * 1024 * 1024, // 10 MB threshold
	}

	// Act
	result, err := guard.ValidateMigration(ctx, sql, opts)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Status != "warn" {
		t.Errorf("expected status 'warn' for large table lock, got '%s'", result.Status)
	}

	hasLockWarning := false
	for _, f := range result.Findings {
		if f.Severity == SeverityWarn && strings.Contains(f.Description, "lock") {
			hasLockWarning = true
			break
		}
	}
	if !hasLockWarning {
		t.Error("expected lock warning for large table")
	}
}

func TestMigrationGuard_ValidateMigration_RollbackAssessment_GeneratesSQL(t *testing.T) {
	// Arrange
	mock := &mockInspector{}
	guard := NewMigrationGuard(mock)
	ctx := context.Background()

	sql := "CREATE TABLE temp_table (id SERIAL PRIMARY KEY)"
	opts := ValidationOptions{
		CheckBreaking:    true,
		CheckPerformance: false,
	}

	// Act
	result, err := guard.ValidateMigration(ctx, sql, opts)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Rollback == nil {
		t.Fatal("expected rollback assessment to be present")
	}
	if !result.Rollback.IsReversible {
		t.Error("expected CREATE TABLE to be reversible")
	}
	if result.Rollback.RollbackSQL == "" {
		t.Error("expected rollback SQL to be generated")
	}
}

func TestMigrationGuard_ValidateMigration_EmptySQL_Passes(t *testing.T) {
	// Arrange
	mock := &mockInspector{}
	guard := NewMigrationGuard(mock)
	ctx := context.Background()

	sql := ""
	opts := ValidationOptions{}

	// Act
	result, err := guard.ValidateMigration(ctx, sql, opts)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Status != "pass" {
		t.Errorf("expected status 'pass' for empty SQL, got '%s'", result.Status)
	}
}

func TestMigrationGuard_ValidateMigration_CheckBreakingDisabled_IgnoresBreaking(t *testing.T) {
	// Arrange
	mock := &mockInspector{}
	guard := NewMigrationGuard(mock)
	ctx := context.Background()

	sql := "DROP TABLE users"
	opts := ValidationOptions{
		CheckBreaking:    false, // Disabled
		CheckPerformance: false,
	}

	// Act
	result, err := guard.ValidateMigration(ctx, sql, opts)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Should not fail even though it's breaking, since check is disabled
	if result.Status == "fail" {
		t.Error("expected not to fail when CheckBreaking is disabled")
	}
}

func TestMigrationGuard_ValidateMigration_MultipleStatements_DropTable_Fails(t *testing.T) {
	mock := &mockInspector{}
	guard := NewMigrationGuard(mock)
	ctx := context.Background()

	sql := "DROP TABLE users; CREATE INDEX idx_users_email ON users (email)"
	opts := ValidationOptions{CheckBreaking: true, CheckPerformance: false}
	res, err := guard.ValidateMigration(ctx, sql, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status != "fail" {
		t.Fatalf("expected fail due to DROP TABLE, got %s", res.Status)
	}
}

func TestMigrationGuard_ValidateMigration_CreateIndexConcurrently_Passes(t *testing.T) {
	mock := &mockInspector{}
	guard := NewMigrationGuard(mock)
	ctx := context.Background()

	sql := "CREATE INDEX CONCURRENTLY idx_orders_user ON orders (user_id)"
	opts := ValidationOptions{CheckBreaking: true, CheckPerformance: false}
	res, err := guard.ValidateMigration(ctx, sql, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status == "fail" {
		t.Fatalf("expected non-fail for create index concurrently")
	}
}

func TestMigrationGuard_ValidateMigration_CreateIndexWithoutConcurrently_Passes(t *testing.T) {
	mock := &mockInspector{}
	guard := NewMigrationGuard(mock)
	ctx := context.Background()

	sql := "CREATE INDEX idx_orders_user ON orders (user_id)"
	opts := ValidationOptions{CheckBreaking: true, CheckPerformance: false}
	res, err := guard.ValidateMigration(ctx, sql, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Status == "fail" {
		t.Fatalf("expected non-fail for create index")
	}
}
