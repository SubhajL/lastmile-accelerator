package storage

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNewPostgresDB_InvalidDSN(t *testing.T) {
	// Malformed DSN returns parse error
	_, err := NewPostgresDB("not-a-valid-dsn")
	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
}

func TestPostgresDB_HealthCheck_Timeout(t *testing.T) {
	// This test requires a real/test DB, so we'll make it integration-ready
	t.Skip("integration test: requires database")

    db, err := NewPostgresDB("pg"+"sql://localhost/nonexistent")
	if err != nil {
		t.Skip("cannot connect to test database")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err = db.HealthCheck(ctx)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestPostgresDB_Close(t *testing.T) {
	// Test that Close doesn't panic on nil or closed connection
	db := &PostgresDB{} // nil DB
	err := db.Close()
	if err != nil {
		t.Errorf("Close on nil DB should not error, got: %v", err)
	}
}

func TestNewPostgresDBFromSQL_WrapsDB(t *testing.T) {
	mockDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mockDB.Close()

	wrapped := NewPostgresDBFromSQL(mockDB)
	if wrapped == nil {
		t.Fatal("expected non-nil wrapped DB")
	}

	if got := wrapped.DB(); got != mockDB {
		t.Errorf("DB() did not return the wrapped *sql.DB: got %p, want %p", got, mockDB)
	}
}
