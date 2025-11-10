package storage

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRunMigrations_ExecutesStatementsInOrder(t *testing.T) {
	// Use sqlmock to assert statements executed
	mdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mdb.Close()

	pdb := NewPostgresDBFromSQL(mdb)

	// Expect CREATE TABLE statements appearing in our migrations
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS project_otel_configs (")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_project_otel_project_id ON project_otel_configs(project_id)")).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS slos (")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_slos_project_id ON slos(project_id)")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS slo_status (")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS slo_history (")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_slo_history_slo_timestamp ON slo_history(slo_id, timestamp DESC)")).WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS alert_rules (")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_alert_rules_slo_id ON alert_rules(slo_id)")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS alert_history (")).WillReturnResult(sqlmock.NewResult(0, 0))
mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_alert_history_rule_timestamp ON alert_history(alert_rule_id, timestamp DESC)")).WillReturnResult(sqlmock.NewResult(0, 0))

// 004 error inbox
mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS error_groups (")).WillReturnResult(sqlmock.NewResult(0,0))
mock.ExpectExec(regexp.QuoteMeta("CREATE UNIQUE INDEX IF NOT EXISTS idx_error_groups_project_fingerprint ON error_groups(project_id, fingerprint)")).WillReturnResult(sqlmock.NewResult(0,0))
mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_error_groups_project_lastseen ON error_groups(project_id, last_seen DESC)")).WillReturnResult(sqlmock.NewResult(0,0))
mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS error_events (")).WillReturnResult(sqlmock.NewResult(0,0))
mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX IF NOT EXISTS idx_error_events_group_ts ON error_events(group_id, ts DESC)")).WillReturnResult(sqlmock.NewResult(0,0))

	if err := RunMigrations(context.Background(), pdb); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
