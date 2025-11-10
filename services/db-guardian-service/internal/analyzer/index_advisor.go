package analyzer

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// IndexAdvisor analyzes query patterns and recommends indexes.
type IndexAdvisor struct {
	inspector DBInspector
}

func NewIndexAdvisor(inspector DBInspector) *IndexAdvisor {
	return &IndexAdvisor{inspector: inspector}
}

// AnalyzeIndexes inspects query stats and existing indexes to recommend new ones.
func (a *IndexAdvisor) AnalyzeIndexes(ctx context.Context, opts IndexAnalysisOptions) (*IndexRecommendations, error) {
	stats, err := a.inspector.GetQueryStats(ctx, opts.MinQueryExecutions)
	if err != nil {
		// Graceful fallback when stats are unavailable (e.g., pg_stat_statements missing)
		stats = nil
	}

	// Build table size map (public schema default)
	tables, _ := a.inspector.GetTables(ctx, "public")
	sizeByTable := map[string]int64{}
	for _, t := range tables {
		sizeByTable[strings.ToLower(t.Name)] = t.SizeBytes
	}

	recs := &IndexRecommendations{AnalyzedAt: time.Now()}
	seen := make(map[string]bool) // key: table|col, used only when CheckDuplicates is true

	for _, s := range stats {
		table, column := extractFilterTarget(strings.ToLower(s.Query))
		if table == "" || column == "" {
			continue
		}

		// Respect table size threshold
		if opts.MinTableSize > 0 {
			if sizeByTable[table] < opts.MinTableSize {
				continue
			}
		}

		// Skip if index already exists
		idxs, _ := a.inspector.GetIndexes(ctx, table)
		if indexExistsOnColumn(idxs, column) && opts.CheckDuplicates {
			continue
		}

		key := table + "|" + column
		if opts.CheckDuplicates {
			if seen[key] {
				continue
			}
			seen[key] = true
		}

		benefit := scoreBenefit(s, sizeByTable[table])
		recs.Recommendations = append(recs.Recommendations, IndexRecommendation{
			TableName:    table,
			Columns:      []string{column},
			Reason:       fmt.Sprintf("Frequent filter on %s.%s without supporting index", table, column),
			BenefitScore: benefit,
			EstimatedImpact: "reduced seq scans and mean latency",
		})
	}


	// Derive duplicates (very simple heuristic)
	if opts.CheckDuplicates {
		recs.Duplicates = findDuplicateIndexes(recs.Recommendations)
	}

	return recs, nil
}

func extractFilterTarget(q string) (table string, column string) {
	// Heuristic parsing: look for "from <table>" and "where <... col ...>"
	fromIdx := strings.Index(q, " from ")
	if fromIdx == -1 {
		return "", ""
	}
	rest := q[fromIdx+6:]
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return "", ""
	}
	table = strings.Trim(parts[0], ",;")

	whereIdx := strings.Index(q, " where ")
	if whereIdx == -1 {
		return table, ""
	}
	cond := q[whereIdx+7:]
	// Try patterns: table.column op or column op
	if idx := strings.Index(cond, table+"."); idx != -1 {
		cond = cond[idx+len(table)+1:]
	}
	// Column name is up to space or operator
	for _, sep := range []string{"=", ">", "<", ">=", "<=", " like ", " ilike ", " in ", ")"} {
		if s := strings.Index(cond, sep); s != -1 {
			column = strings.TrimSpace(cond[:s])
			break
		}
	}
	column = strings.Trim(column, "() ,;")
	return table, column
}

func indexExistsOnColumn(indexes []IndexInfo, column string) bool {
	for _, idx := range indexes {
		for _, c := range idx.Columns {
			if strings.EqualFold(c, column) {
				return true
			}
		}
	}
	return false
}

func scoreBenefit(s QueryStat, tableSize int64) int {
	// Simple scoring: calls weight + size weight
	callsWeight := int(s.Calls / 10)         // each 10 calls adds 1 point
	sizeWeight := int((tableSize / (1024*1024)) / 10) // each 10MB adds 1 point
	if callsWeight < 1 && s.Calls > 0 {
		callsWeight = 1
	}
	return callsWeight + sizeWeight
}

func findDuplicateIndexes(recs []IndexRecommendation) []DuplicateFinding {
	seen := map[string]string{} // key: table|cols -> reason
	var dups []DuplicateFinding
	for _, r := range recs {
		key := r.TableName + "|" + strings.Join(r.Columns, ",")
		if prev, ok := seen[key]; ok {
			dups = append(dups, DuplicateFinding{RedundantIndex: key, CoveredBy: prev, Reason: "duplicate recommendation"})
		} else {
			seen[key] = key
		}
	}
	return dups
}
