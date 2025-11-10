package database

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestWithTx_CommitsOnNilError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectCommit()

	ctx := context.Background()

	// Act
	err = WithTx(ctx, db, func(ctx context.Context, tx Tx) error {
		// Simulate successful operation
		return nil
	})

	// Assert
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestWithTx_RollsBackOnError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectRollback()

	ctx := context.Background()
	testErr := errors.New("operation failed")

	// Act
	err = WithTx(ctx, db, func(ctx context.Context, tx Tx) error {
		// Simulate failed operation
		return testErr
	})

	// Assert
	if err != testErr {
		t.Errorf("expected error %v, got %v", testErr, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestWithTx_ReturnsBeginError(t *testing.T) {
	// Arrange
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}
	defer db.Close()

	beginErr := errors.New("begin failed")
	mock.ExpectBegin().WillReturnError(beginErr)

	ctx := context.Background()

	// Act
	err = WithTx(ctx, db, func(ctx context.Context, tx Tx) error {
		t.Error("function should not be called if begin fails")
		return nil
	})

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
