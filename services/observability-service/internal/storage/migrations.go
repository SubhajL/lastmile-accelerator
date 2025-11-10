package storage

import (
	"context"
	"embed"
	"fmt"
	"regexp"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// RunMigrations executes embedded SQL migrations in lexical order
func RunMigrations(ctx context.Context, pdb *PostgresDB) error {
	files, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	// Ensure deterministic order by filename
	// Read file contents and execute statements sequentially
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		content, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		stmts := splitSQLStatements(string(content))
		for _, stmt := range stmts {
			if strings.TrimSpace(stmt) == "" {
				continue
			}
			if _, err := pdb.db.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("exec migration %s failed: %w", name, err)
			}
		}
	}
	return nil
}

var semicolonSplitter = regexp.MustCompile(`;\s*\n`)

func splitSQLStatements(sql string) []string {
	// Remove line comments
	lines := strings.Split(sql, "\n")
	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "--") {
			lines[i] = ""
		}
	}
	clean := strings.Join(lines, "\n")
	parts := semicolonSplitter.Split(clean, -1)
	return parts
}
