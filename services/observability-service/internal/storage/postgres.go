package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// PostgresDB wraps sql.DB connection pool
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDBFromSQL wraps an existing *sql.DB (used for tests and custom wiring)
func NewPostgresDBFromSQL(db *sql.DB) *PostgresDB {
	return &PostgresDB{db: db}
}

// NewPostgresDB establishes connection pool with pgx driver
func NewPostgresDB(dsn string) (*PostgresDB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)

	// Verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{db: db}, nil
}

// HealthCheck executes SELECT 1 query with 2s timeout
func (p *PostgresDB) HealthCheck(ctx context.Context) error {
	if p.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var result int
	err := p.db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected health check result: %d", result)
	}

	return nil
}

// Close closes connection pool
func (p *PostgresDB) Close() error {
	if p.db == nil {
		return nil
	}
	return p.db.Close()
}

// DB returns underlying sql.DB for queries
func (p *PostgresDB) DB() *sql.DB {
	return p.db
}

// Name returns the dependency name for health checks
func (p *PostgresDB) Name() string {
	return "database"
}
