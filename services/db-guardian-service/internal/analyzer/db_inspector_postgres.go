package analyzer

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// PostgresInspector implements DBInspector for PostgreSQL databases.
type PostgresInspector struct {
	db *sql.DB
}

// NewPostgresInspector creates a new PostgreSQL database inspector.
func NewPostgresInspector(db *sql.DB) *PostgresInspector {
	return &PostgresInspector{db: db}
}

// GetTables retrieves all tables in the specified schema with metadata.
func (i *PostgresInspector) GetTables(ctx context.Context, schema string) ([]TableInfo, error) {
	query := `
		SELECT 
			t.table_schema,
			t.table_name,
			COALESCE(pg_relation_size(quote_ident(t.table_schema)||'.'||quote_ident(t.table_name)), 0) AS size_bytes,
			COALESCE(
				(SELECT reltuples::bigint 
				 FROM pg_class c 
				 JOIN pg_namespace n ON c.relnamespace = n.oid 
				 WHERE n.nspname = t.table_schema AND c.relname = t.table_name),
				0
			) AS row_count
		FROM information_schema.tables t
		WHERE t.table_schema = $1
		  AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_name
	`

	rows, err := i.db.QueryContext(ctx, query, schema)
	if err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo
		if err := rows.Scan(&table.Schema, &table.Name, &table.SizeBytes, &table.RowCount); err != nil {
			return nil, fmt.Errorf("scan table row: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, rows.Err()
}

// GetTableColumns retrieves all columns for a given table.
func (i *PostgresInspector) GetTableColumns(ctx context.Context, tableName string) ([]ColumnInfo, error) {
	query := `
		SELECT 
			column_name,
			data_type,
			is_nullable,
			column_default
		FROM information_schema.columns
		WHERE table_name = $1
		ORDER BY ordinal_position
	`

	rows, err := i.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var nullable string
		var defaultVal sql.NullString

		if err := rows.Scan(&col.Name, &col.DataType, &nullable, &defaultVal); err != nil {
			return nil, fmt.Errorf("scan column row: %w", err)
		}

		col.IsNullable = (nullable == "YES")
		if defaultVal.Valid {
			col.DefaultValue = defaultVal.String
		}

		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// GetIndexes retrieves all indexes for a given table.
func (i *PostgresInspector) GetIndexes(ctx context.Context, tableName string) ([]IndexInfo, error) {
	query := `
		SELECT 
			indexname,
			tablename,
			indexdef
		FROM pg_indexes
		WHERE tablename = $1
		ORDER BY indexname
	`

	rows, err := i.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, fmt.Errorf("query indexes: %w", err)
	}
	defer rows.Close()

	indexes := []IndexInfo{}
	for rows.Next() {
		var idx IndexInfo
		var def string

		if err := rows.Scan(&idx.Name, &idx.TableName, &def); err != nil {
			return nil, fmt.Errorf("scan index row: %w", err)
		}

		// Parse index definition to extract columns
		idx.Columns = parseIndexColumns(def)
		idx.IsUnique = strings.Contains(strings.ToUpper(def), "UNIQUE")

		indexes = append(indexes, idx)
	}

	return indexes, rows.Err()
}

// GetRoles retrieves database roles with their permissions.
func (i *PostgresInspector) GetRoles(ctx context.Context) ([]RoleInfo, error) {
	query := `
		SELECT 
			rolname,
			rolsuper,
			rolinherit,
			rolcreaterole,
			rolcreatedb,
			rolcanlogin
		FROM pg_roles
		WHERE rolname NOT LIKE 'pg_%'
		ORDER BY rolname
	`

	rows, err := i.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	var roles []RoleInfo
	for rows.Next() {
		var role RoleInfo
		var superuser, inherit, createrole, createdb, canlogin bool

		if err := rows.Scan(&role.Name, &superuser, &inherit, &createrole, &createdb, &canlogin); err != nil {
			return nil, fmt.Errorf("scan role row: %w", err)
		}

		// Build permissions list based on flags
		if superuser {
			role.Permissions = append(role.Permissions, Permission{ObjectType: "database", ObjectName: "*", Privilege: "SUPERUSER"})
		}
		if createrole {
			role.Permissions = append(role.Permissions, Permission{ObjectType: "database", ObjectName: "*", Privilege: "CREATEROLE"})
		}
		if createdb {
			role.Permissions = append(role.Permissions, Permission{ObjectType: "database", ObjectName: "*", Privilege: "CREATEDB"})
		}

		role.IsSuperuser = superuser
		role.CanLogin = canlogin
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

// GetQueryStats retrieves query execution statistics (requires pg_stat_statements).
func (i *PostgresInspector) GetQueryStats(ctx context.Context, limit int) ([]QueryStat, error) {
	query := `
		SELECT 
			query,
			calls,
			total_exec_time,
			mean_exec_time
		FROM pg_stat_statements
		ORDER BY calls DESC
		LIMIT $1
	`

	rows, err := i.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("query stats (pg_stat_statements required): %w", err)
	}
	defer rows.Close()

	var stats []QueryStat
	for rows.Next() {
		var stat QueryStat
		if err := rows.Scan(&stat.Query, &stat.Calls, &stat.TotalTime, &stat.MeanTime); err != nil {
			return nil, fmt.Errorf("scan query stat row: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

// AnalyzeMigration parses SQL migration and identifies operations and risks.
func (i *PostgresInspector) AnalyzeMigration(ctx context.Context, sql string) (*MigrationAnalysis, error) {
	analysis := MigrationAnalysis{
		SQL: sql,
	}

	// Normalize SQL for parsing
	normalized := strings.ToUpper(strings.TrimSpace(sql))

	// Detect operations
	operations := detectSQLOperations(normalized)
	analysis.Operations = operations

	// Check for breaking changes
	for _, op := range operations {
		if details, ok := op.Details["is_breaking"]; ok && details == "true" {
			analysis.HasBreaking = true
			break
		}
	}

	return &analysis, nil
}

// parseIndexColumns extracts column names from CREATE INDEX definition.
func parseIndexColumns(def string) []string {
	// Example: CREATE INDEX idx_name ON table_name USING btree (col1, col2)
	start := strings.Index(def, "(")
	end := strings.LastIndex(def, ")")

	if start == -1 || end == -1 || start >= end {
		return []string{}
	}

	columnsStr := def[start+1 : end]
	parts := strings.Split(columnsStr, ",")

	var columns []string
	for _, part := range parts {
		col := strings.TrimSpace(part)
		// Remove any function calls or expressions, keep just column name
		if idx := strings.Index(col, " "); idx != -1 {
			col = col[:idx]
		}
		columns = append(columns, col)
	}

	return columns
}

// detectSQLOperations parses SQL and identifies operation types.
func detectSQLOperations(sql string) []SQLOperation {
	var ops []SQLOperation

	// Split by semicolon for multiple statements
	statements := strings.Split(sql, ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		op := SQLOperation{
			Details: make(map[string]string),
		}

		switch {
		case strings.HasPrefix(stmt, "CREATE TABLE"):
			op.Type = "CREATE"
			op.ObjectType = "TABLE"
			op.Details["lock_level"] = "SHARE"
			op.Details["is_breaking"] = "false"

		case strings.HasPrefix(stmt, "DROP TABLE"):
			op.Type = "DROP"
			op.ObjectType = "TABLE"
			op.Details["lock_level"] = "EXCLUSIVE"
			op.Details["is_breaking"] = "true"

		case strings.HasPrefix(stmt, "ALTER TABLE") && strings.Contains(stmt, "DROP COLUMN"):
			op.Type = "ALTER"
			op.ObjectType = "COLUMN"
			op.Details["lock_level"] = "EXCLUSIVE"
			op.Details["is_breaking"] = "true"
			op.Details["operation"] = "DROP"

		case strings.HasPrefix(stmt, "ALTER TABLE") && strings.Contains(stmt, "ADD COLUMN"):
			op.Type = "ALTER"
			op.ObjectType = "COLUMN"
			op.Details["lock_level"] = "SHARE"
			op.Details["is_breaking"] = "false"
			op.Details["operation"] = "ADD"

		case strings.HasPrefix(stmt, "CREATE INDEX"):
			op.Type = "CREATE"
			op.ObjectType = "INDEX"
			if strings.Contains(stmt, "CONCURRENTLY") {
				op.Details["lock_level"] = "SHARE UPDATE EXCLUSIVE"
			} else {
				op.Details["lock_level"] = "SHARE"
			}
			op.Details["is_breaking"] = "false"

		case strings.HasPrefix(stmt, "DROP INDEX"):
			op.Type = "DROP"
			op.ObjectType = "INDEX"
			op.Details["lock_level"] = "EXCLUSIVE"
			op.Details["is_breaking"] = "false"

		default:
			op.Type = "UNKNOWN"
			op.ObjectType = "UNKNOWN"
			op.Details["lock_level"] = "UNKNOWN"
			op.Details["is_breaking"] = "false"
		}

		ops = append(ops, op)
	}

	return ops
}
