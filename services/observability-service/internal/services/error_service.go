package services

import (
	"context"

	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/repository"
	"go.opentelemetry.io/otel/trace"
)

type ErrorService interface {
	Ingest(ctx context.Context, projectID, message, stack, fingerprint, title string, metadata map[string]interface{}) (string, error)
	ListGroups(ctx context.Context, projectID string, filter models.ErrorGroupFilter) ([]models.ErrorGroup, error)
	GetGroup(ctx context.Context, groupID string) (*models.ErrorGroup, error)
	ListGroupEvents(ctx context.Context, groupID string, limit, offset int) ([]models.ErrorEvent, error)
	ResolveGroup(ctx context.Context, groupID string) error
}

type errorService struct {
	repo   repository.ErrorRepository
	tracer trace.Tracer
}

func NewErrorService(repo repository.ErrorRepository, tracer trace.Tracer) ErrorService {
	if tracer == nil { tracer = trace.NewNoopTracerProvider().Tracer("error-service") }
	return &errorService{repo: repo, tracer: tracer}
}

func (s *errorService) Ingest(ctx context.Context, projectID, message, stack, fingerprint, title string, metadata map[string]interface{}) (string, error) {
	_, span := s.tracer.Start(ctx, "error.ingest")
	defer span.End()
	return s.repo.RecordEvent(ctx, projectID, fingerprint, title, message, stack, metadata)
}

func (s *errorService) ListGroups(ctx context.Context, projectID string, filter models.ErrorGroupFilter) ([]models.ErrorGroup, error) {
	return s.repo.ListGroups(ctx, projectID, filter)
}

func (s *errorService) GetGroup(ctx context.Context, groupID string) (*models.ErrorGroup, error) {
	return s.repo.GetGroup(ctx, groupID)
}

func (s *errorService) ListGroupEvents(ctx context.Context, groupID string, limit, offset int) ([]models.ErrorEvent, error) {
	return s.repo.ListEvents(ctx, groupID, limit, offset)
}

func (s *errorService) ResolveGroup(ctx context.Context, groupID string) error { return s.repo.ResolveGroup(ctx, groupID) }
