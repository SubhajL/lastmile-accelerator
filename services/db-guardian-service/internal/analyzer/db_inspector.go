package analyzer

import "context"

// DBInspector is an interface for database introspection operations
// This abstraction allows testing analyzers without real databases
type DBInspector interface {
	// GetTables lists all tables in schema with row counts and sizes
	// Returns empty slice if schema doesn't exist
	GetTables(ctx context.Context, schema string) ([]TableInfo, error)

	// GetTableColumns describes columns for a table
	// Returns error if table not found
	GetTableColumns(ctx context.Context, table string) ([]ColumnInfo, error)

	// GetIndexes lists existing indexes on table
	// Returns empty slice if no indexes
	GetIndexes(ctx context.Context, table string) ([]IndexInfo, error)

	// GetRoles lists database roles with granted permissions
	// Returns all roles visible to connection user
	GetRoles(ctx context.Context) ([]RoleInfo, error)

	// GetQueryStats fetches query statistics for frequently executed queries
	// Uses pg_stat_statements if available
	GetQueryStats(ctx context.Context, minExecutions int) ([]QueryStat, error)

	// AnalyzeMigration dry-run analyzes migration SQL
	// Returns warnings/errors without applying changes
	AnalyzeMigration(ctx context.Context, sql string) (*MigrationAnalysis, error)
}
