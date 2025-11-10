package analyzer

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// MigrationGuard validates database migrations for safety and performance.
type MigrationGuard struct {
	inspector DBInspector
}

// NewMigrationGuard creates a new migration guard.
func NewMigrationGuard(inspector DBInspector) *MigrationGuard {
	return &MigrationGuard{inspector: inspector}
}

// ValidateMigration analyzes migration SQL and returns validation result.
func (g *MigrationGuard) ValidateMigration(ctx context.Context, sql string, opts ValidationOptions) (*MigrationValidationResult, error) {
	result := &MigrationValidationResult{
		Status:      "pass",
		Findings:    []Finding{},
		ValidatedAt: time.Now(),
	}

	// Handle empty SQL
	if strings.TrimSpace(sql) == "" {
		return result, nil
	}

	// Analyze migration using inspector
	analysis, err := g.inspector.AnalyzeMigration(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("analyze migration: %w", err)
	}

	// Check for breaking changes
	if opts.CheckBreaking && analysis.HasBreaking {
		result.Status = "fail"
		result.Findings = append(result.Findings, Finding{
			Severity:    SeverityCritical,
			Category:    "breaking_change",
			Title:       "Breaking schema change detected",
			Description: "Migration contains operations that may break existing application code",
			Remediation: "Review breaking operations. Consider backwards-compatible migrations with staged rollout.",
		})
	}

	// Check performance implications
	if opts.CheckPerformance {
		performanceFindings := g.checkPerformanceImpact(ctx, analysis, opts)
		result.Findings = append(result.Findings, performanceFindings...)

		// Upgrade status to warn if we have warnings
		for _, f := range performanceFindings {
			if f.Severity == SeverityWarn && result.Status == "pass" {
				result.Status = "warn"
			}
		}
	}

	// Generate rollback assessment
	result.Rollback = g.generateRollbackAssessment(analysis)

	return result, nil
}

// checkPerformanceImpact analyzes migration for performance issues.
func (g *MigrationGuard) checkPerformanceImpact(ctx context.Context, analysis *MigrationAnalysis, opts ValidationOptions) []Finding {
	var findings []Finding

	// Extract affected tables
	affectedTables := g.extractAffectedTables(analysis)

	// Check table sizes for lock warnings
	for _, tableName := range affectedTables {
		tables, err := g.inspector.GetTables(ctx, "public")
		if err != nil {
			continue // Gracefully skip if can't fetch tables
		}

		for _, table := range tables {
			if table.Name == tableName && table.SizeBytes > opts.MaxTableSize {
				findings = append(findings, Finding{
					Severity:    SeverityWarn,
					Category:    "performance",
					Title:       "Migration may cause extended table lock",
					Description: fmt.Sprintf("Table '%s' is %.2f MB. Migration may hold locks and impact availability.", tableName, float64(table.SizeBytes)/(1024*1024)),
					Remediation: "Consider using CONCURRENTLY for index operations or splitting migration into smaller batches.",
				})
			}
		}
	}

	return findings
}

// extractAffectedTables gets table names from operations.
func (g *MigrationGuard) extractAffectedTables(analysis *MigrationAnalysis) []string {
	tableMap := make(map[string]bool)

	for _, op := range analysis.Operations {
		if op.ObjectName != "" {
			tableMap[op.ObjectName] = true
		}
	}

	// Extract from SQL
	sql := strings.ToUpper(analysis.SQL)
	
	// Look for "ALTER TABLE tablename" pattern
	if strings.Contains(sql, "ALTER TABLE") {
		if idx := strings.Index(sql, "ALTER TABLE "); idx != -1 {
			rest := sql[idx+12:] // len("ALTER TABLE ") = 12
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				tableName := strings.TrimSuffix(parts[0], ";")
				tableName = strings.TrimSuffix(tableName, ",")
				tableName = strings.ToLower(tableName)
				tableMap[tableName] = true
			}
		}
	}
	
	// Look for "CREATE TABLE tablename" pattern
	if strings.Contains(sql, "CREATE TABLE") {
		if idx := strings.Index(sql, "CREATE TABLE "); idx != -1 {
			rest := sql[idx+13:] // len("CREATE TABLE ") = 13
			parts := strings.Fields(rest)
			if len(parts) > 0 {
				tableName := strings.TrimSuffix(parts[0], ";")
				tableName = strings.TrimSuffix(tableName, ",")
				tableName = strings.TrimSuffix(tableName, "(")
				tableName = strings.ToLower(tableName)
				tableMap[tableName] = true
			}
		}
	}

	var tables []string
	for table := range tableMap {
		tables = append(tables, table)
	}
	return tables
}

// generateRollbackAssessment evaluates whether migration can be rolled back.
func (g *MigrationGuard) generateRollbackAssessment(analysis *MigrationAnalysis) *RollbackAssessment {
	assessment := &RollbackAssessment{
		IsReversible: true,
		Notes:        "",
	}

	var rollbackStatements []string

	for _, op := range analysis.Operations {
		switch op.Type {
		case "CREATE":
			if op.ObjectType == "TABLE" {
				assessment.IsReversible = true
				rollbackStatements = append(rollbackStatements, fmt.Sprintf("DROP TABLE IF EXISTS %s;", op.ObjectName))
			} else if op.ObjectType == "INDEX" {
				assessment.IsReversible = true
				rollbackStatements = append(rollbackStatements, fmt.Sprintf("DROP INDEX IF EXISTS %s;", op.ObjectName))
			}

		case "DROP":
			assessment.IsReversible = false
			assessment.Notes += "DROP operations cannot be automatically reversed. "

		case "ALTER":
			if op.Details["operation"] == "DROP" {
				assessment.IsReversible = false
				assessment.Notes += "Column drops cannot be automatically reversed. "
			} else {
				assessment.IsReversible = true
				assessment.Notes += "ALTER operations may require manual rollback. "
			}

		default:
			assessment.Notes += "Unknown operation type, manual review required. "
		}
	}

	if len(rollbackStatements) > 0 {
		assessment.RollbackSQL = strings.Join(rollbackStatements, "\n")
	}

	return assessment
}
