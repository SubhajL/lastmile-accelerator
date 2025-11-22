package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"example.com/lma/db-guardian-service/internal/analyzer"
	"example.com/lma/db-guardian-service/internal/models"
	"example.com/lma/db-guardian-service/internal/repository"
	"example.com/lma/db-guardian-service/internal/events"
	"github.com/nats-io/nats.go"
)

// AnalysisReport aggregates outputs of all analyzers.
type AnalysisReport struct {
	Role      *analyzer.RoleAnalysisResult
	Migration *analyzer.MigrationValidationResult
	Index     *analyzer.IndexRecommendations
}

// ProjectAnalysisRequest contains the request parameters for project analysis
type ProjectAnalysisRequest struct {
	ProjectID     string
	MigrationName string
	MigrationSQL  string
	RoleOpts      analyzer.AnalyzeOptions
	ValOpts       analyzer.ValidationOptions
	IdxOpts       analyzer.IndexAnalysisOptions
}

// ProjectDatabaseResolver defines the interface for resolving project database connections
type ProjectDatabaseResolver interface {
	ResolveInspector(ctx context.Context, projectID string) (analyzer.DBInspector, func() error, error)
}

// AnalysisService orchestrates analyzers and persists results.
type EventPublisher func(ctx context.Context, nc *nats.Conn, subject string, data []byte) error

type AnalysisService struct {
	db       *sql.DB              // Service metadata DB (unchanged)
	nats     *nats.Conn
	resolver ProjectDatabaseResolver     // NEW

	// Remove these (will be created per-request):
	// roles      *analyzer.RoleAnalyzer
	// migrations *analyzer.MigrationGuard
	// indexes    *analyzer.IndexAdvisor

	publish EventPublisher

	// Add analyzer factories:
	newRoleAnalyzer      func(analyzer.DBInspector) *analyzer.RoleAnalyzer
	newMigrationGuard    func(analyzer.DBInspector) *analyzer.MigrationGuard
	newIndexAdvisor      func(analyzer.DBInspector) *analyzer.IndexAdvisor
}

func NewAnalysisService(db *sql.DB, natsConn *nats.Conn, resolver ProjectDatabaseResolver) *AnalysisService {
	s := &AnalysisService{
		db:       db,
		nats:     natsConn,
		resolver: resolver,
	}
	s.publish = events.PublishEvent

	// Set default analyzer factories
	s.newRoleAnalyzer = func(inspector analyzer.DBInspector) *analyzer.RoleAnalyzer {
		return analyzer.NewRoleAnalyzer(inspector)
	}
	s.newMigrationGuard = func(inspector analyzer.DBInspector) *analyzer.MigrationGuard {
		return analyzer.NewMigrationGuard(inspector)
	}
	s.newIndexAdvisor = func(inspector analyzer.DBInspector) *analyzer.IndexAdvisor {
		return analyzer.NewIndexAdvisor(inspector)
	}

	return s
}

// SetPublisher allows injecting a custom event publisher (for testing)
func (s *AnalysisService) SetPublisher(pub EventPublisher) {
	s.publish = pub
}

// RunProjectAnalysis runs analysis using project-scoped database connection
func (s *AnalysisService) RunProjectAnalysis(ctx context.Context, req ProjectAnalysisRequest) (*AnalysisReport, error) {
	// Validate request
	if req.ProjectID == "" {
		return nil, fmt.Errorf("projectID is required")
	}

	// Resolve project database and get inspector
	inspector, closer, err := s.resolver.ResolveInspector(ctx, req.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project database: %w", err)
	}
	// Ensure cleanup even on panic/error
	defer closer()

	// Create project-scoped analyzers
	roles := s.newRoleAnalyzer(inspector)
	migrations := s.newMigrationGuard(inspector)
	indexes := s.newIndexAdvisor(inspector)

	// Execute analyzers
	roleRes, err := roles.Analyze(ctx, req.RoleOpts)
	if err != nil {
		return nil, fmt.Errorf("role analysis failed: %w", err)
	}

	migRes, err := migrations.ValidateMigration(ctx, req.MigrationSQL, req.ValOpts)
	if err != nil {
		return nil, fmt.Errorf("migration validation failed: %w", err)
	}

	idxRes, err := indexes.AnalyzeIndexes(ctx, req.IdxOpts)
	if err != nil {
		return nil, fmt.Errorf("index analysis failed: %w", err)
	}

	// Persist migration audit to service DB
	if s.db != nil {
		migRepo := repository.NewMigrationAuditsRepository(s.db)
		findingsBytes, _ := json.Marshal(migRes.Findings)
		_, _ = migRepo.Record(ctx, &models.MigrationAudit{
			ProjectID:     req.ProjectID,
			MigrationName: req.MigrationName,
			Status:        migRes.Status,
			FindingsJSON:  string(findingsBytes),
		})
	}

	// Persist index recommendations to service DB
	if s.db != nil {
		idxRepo := repository.NewIndexRecommendationsRepository(s.db)
		for _, r := range idxRes.Recommendations {
			_, _ = idxRepo.Upsert(ctx, &models.IndexRecommendation{
				ProjectID:    req.ProjectID,
				TableName:    r.TableName,
				ColumnNames:  r.Columns,
				Reason:       r.Reason,
				BenefitScore: r.BenefitScore,
				Applied:      false,
			})
		}
	}

	// Publish events if NATS configured (sanitized, no DSN)
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

// RunFullAnalysis - DEPRECATED: Use RunProjectAnalysis instead
func (s *AnalysisService) RunFullAnalysis(
	ctx context.Context,
	projectID string,
	migrationName string,
	migrationSQL string,
	roleOpts analyzer.AnalyzeOptions,
	valOpts analyzer.ValidationOptions,
	idxOpts analyzer.IndexAnalysisOptions,
) (*AnalysisReport, error) {
	// Delegate to the new method
	return s.RunProjectAnalysis(ctx, ProjectAnalysisRequest{
		ProjectID:     projectID,
		MigrationName: migrationName,
		MigrationSQL:  migrationSQL,
		RoleOpts:      roleOpts,
		ValOpts:       valOpts,
		IdxOpts:       idxOpts,
	})
}
