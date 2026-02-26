package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/events"
	"example.com/lma/db-guardian-service/internal/models"
	"example.com/lma/db-guardian-service/internal/repository"
	"github.com/nats-io/nats.go"
)

// AnalysisReport aggregates outputs of all analyzers.
type AnalysisReport struct {
	Role      *analyzer.RoleAnalysisResult
	Migration *analyzer.MigrationValidationResult
	Index     *analyzer.IndexRecommendations
}

// AnalysisService orchestrates analyzers and persists results.
type EventPublisher func(ctx context.Context, nc *nats.Conn, subject string, data []byte) error

type AnalysisService struct {
	db         *sql.DB
	nats       *nats.Conn
	roles      *analyzer.RoleAnalyzer
	migrations *analyzer.MigrationGuard
	indexes    *analyzer.IndexAdvisor
	publish    EventPublisher
	resolver   InspectorResolver
}

func NewAnalysisService(db *sql.DB, natsConn *nats.Conn, role *analyzer.RoleAnalyzer, guard *analyzer.MigrationGuard, idx *analyzer.IndexAdvisor) *AnalysisService {
	s := &AnalysisService{db: db, nats: natsConn, roles: role, migrations: guard, indexes: idx}
	s.publish = events.PublishEvent
	return s
}

type InspectorResolver interface {
	ResolveInspector(ctx context.Context, projectID string) (analyzer.DBInspector, func() error, error)
}

func NewAnalysisServiceWithProjectResolver(db *sql.DB, natsConn *nats.Conn, resolver InspectorResolver) *AnalysisService {
	s := &AnalysisService{db: db, nats: natsConn, resolver: resolver}
	s.publish = events.PublishEvent
	return s
}

// SetPublisher allows injecting a custom event publisher (for testing)
func (s *AnalysisService) SetPublisher(pub EventPublisher) {
	s.publish = pub
}

// RunFullAnalysis runs role analysis, migration validation, and index advisory for a project, then persists and publishes events.
func (s *AnalysisService) RunFullAnalysis(
	ctx context.Context,
	projectID string,
	migrationName string,
	migrationSQL string,
	roleOpts analyzer.AnalyzeOptions,
	valOpts analyzer.ValidationOptions,
	idxOpts analyzer.IndexAnalysisOptions,
) (*AnalysisReport, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}

	roles := s.roles
	migrations := s.migrations
	indexes := s.indexes
	if s.resolver != nil {
		inspector, cleanup, err := s.resolver.ResolveInspector(ctx, projectID)
		if err != nil {
			return nil, err
		}
		if cleanup != nil {
			defer func() {
				if err := cleanup(); err != nil {
					log.Printf("project DB cleanup error: %v", err)
				}
			}()
		}
		roles = analyzer.NewRoleAnalyzer(inspector)
		migrations = analyzer.NewMigrationGuard(inspector)
		indexes = analyzer.NewIndexAdvisor(inspector)
	}

	if roles == nil || migrations == nil || indexes == nil {
		return nil, fmt.Errorf("analysis service not configured")
	}

	// Execute analyzers
	roleRes, err := roles.Analyze(ctx, roleOpts)
	if err != nil {
		return nil, fmt.Errorf("role analysis failed: %w", err)
	}

	migRes, err := migrations.ValidateMigration(ctx, migrationSQL, valOpts)
	if err != nil {
		return nil, fmt.Errorf("migration validation failed: %w", err)
	}

	idxRes, err := indexes.AnalyzeIndexes(ctx, idxOpts)
	if err != nil {
		return nil, fmt.Errorf("index analysis failed: %w", err)
	}

	// Persist migration audit
	if s.db != nil {
		migRepo := repository.NewMigrationAuditsRepository(s.db)
		findingsBytes, _ := json.Marshal(migRes.Findings)
		_, _ = migRepo.Record(ctx, &models.MigrationAudit{
			ProjectID:     projectID,
			MigrationName: migrationName,
			Status:        migRes.Status,
			FindingsJSON:  string(findingsBytes),
		})
	}

	// Persist index recommendations
	if s.db != nil {
		idxRepo := repository.NewIndexRecommendationsRepository(s.db)
		for _, r := range idxRes.Recommendations {
			_, _ = idxRepo.Upsert(ctx, &models.IndexRecommendation{
				ProjectID:    projectID,
				TableName:    r.TableName,
				ColumnNames:  r.Columns,
				Reason:       r.Reason,
				BenefitScore: r.BenefitScore,
				Applied:      false,
			})
		}
	}

	// Publish events if NATS configured
	if s.nats != nil && s.publish != nil {
		// analysis completed
		payload, _ := json.Marshal(roleRes)
		_ = s.publish(ctx, s.nats, "db.analysis.completed", payload)

		payload, _ = json.Marshal(migRes)
		_ = s.publish(ctx, s.nats, "db.migration.validated", payload)

		payload, _ = json.Marshal(idxRes)
		_ = s.publish(ctx, s.nats, "db.index.recommended", payload)
	}

	return &AnalysisReport{Role: roleRes, Migration: migRes, Index: idxRes}, nil
}
