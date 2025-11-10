package services

import (
	"context"
	"fmt"
	"time"

	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/repository"
	"go.opentelemetry.io/otel/trace"
)

type OTelService interface {
	GetAvailablePresets(ctx context.Context) ([]*models.OTelPreset, error)
	GetPresetForFramework(ctx context.Context, framework models.Framework) (*models.OTelPreset, error)
	ApplyPresetToProject(ctx context.Context, projectID string, framework models.Framework) (*models.ProjectOTelConfig, error)
	GetProjectConfiguration(ctx context.Context, projectID string) (*models.ProjectOTelConfig, error)
}

type otelService struct {
	repo   repository.OTelRepository
	tracer trace.Tracer
}

func NewOTelService(repo repository.OTelRepository, tracer trace.Tracer) OTelService {
	if tracer == nil {
		tracer = trace.NewNoopTracerProvider().Tracer("otel-service")
	}
	return &otelService{repo: repo, tracer: tracer}
}

func (s *otelService) GetAvailablePresets(ctx context.Context) ([]*models.OTelPreset, error) {
	_, span := s.tracer.Start(ctx, "otel.get_presets")
	defer span.End()
	return s.repo.ListPresets(ctx)
}

func (s *otelService) GetPresetForFramework(ctx context.Context, framework models.Framework) (*models.OTelPreset, error) {
	_, span := s.tracer.Start(ctx, "otel.get_preset")
	defer span.End()
	if !models.IsValidFramework(framework) {
		return nil, fmt.Errorf("invalid framework: %s", framework)
	}
	return s.repo.GetPreset(ctx, framework)
}

func (s *otelService) ApplyPresetToProject(ctx context.Context, projectID string, framework models.Framework) (*models.ProjectOTelConfig, error) {
	_, span := s.tracer.Start(ctx, "otel.apply_preset")
	defer span.End()
	preset, err := s.repo.GetPreset(ctx, framework)
	if err != nil {
		return nil, err
	}
	cfg := &models.ProjectOTelConfig{
		ID:        newUUID(),
		ProjectID: projectID,
		Framework: framework,
		Config: map[string]interface{}{
			"trace_endpoint":  preset.TraceEndpoint,
			"metric_endpoint": preset.MetricEndpoint,
			"sampling_rate":   preset.SamplingRate,
			"exporter_type":   preset.ExporterType,
		},
		AppliedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.repo.SaveProjectConfig(ctx, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *otelService) GetProjectConfiguration(ctx context.Context, projectID string) (*models.ProjectOTelConfig, error) {
	_, span := s.tracer.Start(ctx, "otel.get_project_config")
	defer span.End()
	if projectID == "" {
		return nil, fmt.Errorf("projectID required")
	}
	return s.repo.GetProjectConfig(ctx, projectID)
}

// newUUID is a very small helper for tests; replace with proper uuid lib later
func newUUID() string { return fmt.Sprintf("%d", time.Now().UnixNano()) }
