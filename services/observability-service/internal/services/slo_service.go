package services

import (
	"context"
	"fmt"
	"time"

	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/repository"
	"go.opentelemetry.io/otel/trace"
)

type MetricClient interface {
	Query(ctx context.Context, promQL string, window time.Duration) (float64, error)
}

type SLOService interface {
	CreateSLO(ctx context.Context, slo *models.SLO) error
	GetSLO(ctx context.Context, id string) (*models.SLO, error)
	ListProjectSLOs(ctx context.Context, projectID string) ([]*models.SLO, error)
	UpdateSLO(ctx context.Context, slo *models.SLO) error
	DeleteSLO(ctx context.Context, id string) error
	EvaluateSLO(ctx context.Context, sloID string) (*models.SLOStatus, error)
	GetSLOStatus(ctx context.Context, sloID string) (*models.SLOStatus, error)
	GetSLOHistory(ctx context.Context, sloID string, from, to time.Time) ([]*models.SLOHistory, error)
}

type sloService struct {
	repo   repository.SLORepository
	mc     MetricClient
	tracer trace.Tracer
}

func NewSLOService(repo repository.SLORepository, tracer trace.Tracer, mc MetricClient) SLOService {
	if tracer == nil {
		tracer = trace.NewNoopTracerProvider().Tracer("slo-service")
	}
	return &sloService{repo: repo, tracer: tracer, mc: mc}
}

func (s *sloService) CreateSLO(ctx context.Context, slo *models.SLO) error {
	if err := slo.Validate(); err != nil {
		return err
	}
	return s.repo.Create(ctx, slo)
}

func (s *sloService) GetSLO(ctx context.Context, id string) (*models.SLO, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *sloService) ListProjectSLOs(ctx context.Context, projectID string) ([]*models.SLO, error) {
	return s.repo.ListByProject(ctx, projectID)
}
func (s *sloService) UpdateSLO(ctx context.Context, slo *models.SLO) error {
	if err := slo.Validate(); err != nil {
		return err
	}
	return s.repo.Update(ctx, slo)
}
func (s *sloService) DeleteSLO(ctx context.Context, id string) error { return s.repo.Delete(ctx, id) }

func (s *sloService) EvaluateSLO(ctx context.Context, sloID string) (*models.SLOStatus, error) {
	slo, err := s.repo.GetByID(ctx, sloID)
	if err != nil {
		return nil, err
	}
	actual, err := s.mc.Query(ctx, slo.Query, slo.Window)
	if err != nil {
		return nil, fmt.Errorf("metric query: %w", err)
	}
	comp := s.calculateCompliance(slo.Target, actual, slo.Type)
	status := &models.SLOStatus{
		SLOID:           sloID,
		Compliance:      comp,
		RemainingBudget: 100 - comp,
		BurnRate:        s.calculateBurnRate(comp, comp, slo.Window),
		LastCalculated:  time.Now(),
	}
	if err := s.repo.SaveStatus(ctx, status); err != nil {
		return nil, err
	}
	return status, nil
}

func (s *sloService) GetSLOStatus(ctx context.Context, sloID string) (*models.SLOStatus, error) {
	return s.repo.GetStatus(ctx, sloID)
}

func (s *sloService) GetSLOHistory(ctx context.Context, sloID string, from, to time.Time) ([]*models.SLOHistory, error) {
	return s.repo.GetHistory(ctx, sloID, from, to)
}

func (s *sloService) calculateCompliance(target float64, actual float64, sloType models.SLOType) float64 {
	switch sloType {
	case models.SLOTypeAvailability, models.SLOTypeThroughput:
		if target == 0 {
			return 0
		}
		return min(100, (actual/target)*100)
	case models.SLOTypeLatency, models.SLOTypeErrorRate:
		if target == 0 {
			return 0
		}
		return min(100, (target/actual)*100)
	default:
		return actual
	}
}

func (s *sloService) calculateBurnRate(current, previous float64, window time.Duration) float64 {
	// Simple derivative per hour
	if window.Hours() == 0 {
		return 0
	}
	return (previous - current) / window.Hours()
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
