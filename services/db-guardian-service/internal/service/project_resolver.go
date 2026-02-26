package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/models"
)

type DefaultConnectionGetter interface {
	GetDefaultConnection(ctx context.Context, projectID string) (*models.DBConnection, error)
}

type ProjectDSNGetter interface {
	GetProjectDSN(ctx context.Context, path string) (string, error)
}

type DBOpener func(ctx context.Context, dsn string) (*sql.DB, error)

type ProjectResolver struct {
	connections DefaultConnectionGetter
	vault       ProjectDSNGetter
	openDB      DBOpener
}

func NewProjectResolver(
	connections DefaultConnectionGetter,
	vault ProjectDSNGetter,
	openDB DBOpener,
) *ProjectResolver {
	return &ProjectResolver{
		connections: connections,
		vault:       vault,
		openDB:      openDB,
	}
}

func (r *ProjectResolver) ResolveInspector(
	ctx context.Context,
	projectID string,
) (analyzer.DBInspector, func() error, error) {
	if strings.TrimSpace(projectID) == "" {
		return nil, nil, fmt.Errorf("projectID is required")
	}
	if r == nil || r.connections == nil {
		return nil, nil, fmt.Errorf("connections dependency not configured")
	}
	if r.vault == nil {
		return nil, nil, fmt.Errorf("vault dependency not configured")
	}
	if r.openDB == nil {
		return nil, nil, fmt.Errorf("db opener dependency not configured")
	}

	conn, err := r.connections.GetDefaultConnection(ctx, projectID)
	if err != nil {
		return nil, nil, fmt.Errorf("get default connection: %w", err)
	}
	if conn == nil {
		return nil, nil, fmt.Errorf("default connection not found")
	}
	if !strings.EqualFold(conn.Driver, "postgres") {
		return nil, nil, fmt.Errorf("unsupported driver: %s", conn.Driver)
	}
	if strings.TrimSpace(conn.DSNRef) == "" {
		return nil, nil, fmt.Errorf("dsn_ref is required")
	}

	dsn, err := r.vault.GetProjectDSN(ctx, conn.DSNRef)
	if err != nil {
		return nil, nil, fmt.Errorf("get project DSN: %w", err)
	}

	db, err := r.openDB(ctx, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("open project DB: %w", err)
	}

	return analyzer.NewPostgresInspector(db), db.Close, nil
}
