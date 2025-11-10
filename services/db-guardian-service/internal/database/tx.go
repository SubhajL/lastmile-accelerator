package database

import (
	"context"
	"database/sql"
	"fmt"
)

// Tx is a narrow interface for database transactions
// This allows for easier testing and abstraction
type Tx interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// WithTx executes a function within a database transaction
// The transaction is committed if the function returns nil
// The transaction is rolled back if the function returns an error
func WithTx(ctx context.Context, db *sql.DB, fn func(context.Context, Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute the function
	err = fn(ctx, tx)
	if err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %w (original error: %v)", rbErr, err)
		}
		return err
	}

	// Commit on success
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
