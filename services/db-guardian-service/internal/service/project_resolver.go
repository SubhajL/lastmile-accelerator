package service

import (
	"context"
	"database/sql"
	"fmt"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/repository"
	"example.com/lma/db-guardian-service/internal/secrets"
)

// VaultDSNFetcher defines the interface for fetching DSNs from Vault
type VaultDSNFetcher interface {
	GetProjectDSN(ctx context.Context, vaultRef string) (string, error)
}

// ProjectResolver resolves project databases by fetching connection info from metadata DB
// and actual DSNs from Vault, then opens temporary DB connections for analysis.
type ProjectResolver struct {
	db           *sql.DB                              // Service metadata DB
	vault        VaultDSNFetcher
	connRepo     *repository.ConnectionsRepository
	inspectorNew func(*sql.DB) analyzer.DBInspector   // Factory for testing
}

// NewProjectResolver creates a new project database resolver.
func NewProjectResolver(db *sql.DB, vault *secrets.VaultClient) *ProjectResolver {
	return &ProjectResolver{
		db:           db,
		vault:        vault,
		connRepo:     repository.NewConnectionsRepository(db),
		inspectorNew: func(db *sql.DB) analyzer.DBInspector {
			return analyzer.NewPostgresInspector(db)
		},
	}
}

// Resolve resolves a project database connection.
// Returns the opened database connection and a cleanup function that must be called.
// The caller is responsible for calling the cleanup function to close the connection.
func (r *ProjectResolver) Resolve(ctx context.Context, projectID string) (*sql.DB, func() error, error) {
	// Validate projectID
	if projectID == "" {
		return nil, nil, fmt.Errorf("projectID cannot be empty")
	}

	// Get connection metadata from repository
	conn, err := r.connRepo.GetDefaultByProject(ctx, projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get connection for project %s: %w", projectID, err)
	}

	// Fetch actual DSN from Vault
	dsn, err := r.vault.GetProjectDSN(ctx, conn.DSNRef)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch DSN for project %s: %w", projectID, err)
	}

	// Validate driver
	if conn.Driver == "" {
		return nil, nil, fmt.Errorf("invalid database driver: driver cannot be empty")
	}

	// Open database connection
	db, err := sql.Open(conn.Driver, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Note: We're not pinging the database here to verify connectivity.
	// sql.Open doesn't actually establish a connection; it just validates arguments.
	// The actual connection will be established on first use.
	// This approach allows for easier testing and follows the lazy connection pattern.

	// Return db and cleanup function
	return db, db.Close, nil
}

// ResolveInspector resolves a project database and creates a DBInspector for it.
// Returns the inspector and a cleanup function that must be called.
// The caller is responsible for calling the cleanup function to close the underlying connection.
func (r *ProjectResolver) ResolveInspector(ctx context.Context, projectID string) (analyzer.DBInspector, func() error, error) {
	// Resolve database connection
	db, closer, err := r.Resolve(ctx, projectID)
	if err != nil {
		return nil, nil, err
	}

	// Create inspector
	inspector := r.inspectorNew(db)

	// Return inspector and closer
	return inspector, closer, nil
}