package services

import (
	"context"
	"testing"
	"time"

	"example.com/lma/observability-service/internal/models"
	"go.opentelemetry.io/otel/trace"
)

type fakeSLORepo struct {
	m      map[string]*models.SLO
	status *models.SLOStatus
}

func (f *fakeSLORepo) Create(ctx context.Context, slo *models.SLO) error {
	f.m[slo.ID] = slo
	return nil
}
func (f *fakeSLORepo) GetByID(ctx context.Context, id string) (*models.SLO, error) {
	return f.m[id], nil
}
func (f *fakeSLORepo) ListByProject(ctx context.Context, projectID string) ([]*models.SLO, error) {
	return nil, nil
}
func (f *fakeSLORepo) ListAll(ctx context.Context) ([]*models.SLO, error) { return nil, nil }
func (f *fakeSLORepo) Update(ctx context.Context, slo *models.SLO) error {
	f.m[slo.ID] = slo
	return nil
}
func (f *fakeSLORepo) Delete(ctx context.Context, id string) error { delete(f.m, id); return nil }
func (f *fakeSLORepo) SaveStatus(ctx context.Context, status *models.SLOStatus) error {
	f.status = status
	return nil
}
func (f *fakeSLORepo) GetStatus(ctx context.Context, sloID string) (*models.SLOStatus, error) {
	return f.status, nil
}
func (f *fakeSLORepo) GetHistory(ctx context.Context, sloID string, from, to time.Time) ([]*models.SLOHistory, error) {
	return nil, nil
}

type fakeMetricClient struct{ val float64 }

func (f *fakeMetricClient) Query(ctx context.Context, promQL string, window time.Duration) (float64, error) {
	return f.val, nil
}

func TestSLOService_EvaluateAndCalculations(t *testing.T) {
	repo := &fakeSLORepo{m: map[string]*models.SLO{}}
	var noop trace.Tracer
	mc := &fakeMetricClient{val: 99.0}
	svc := NewSLOService(repo, noop, mc)

	slo := &models.SLO{ID: "s1", ProjectID: "p1", ServiceName: "svc", Type: models.SLOTypeAvailability, Target: 99.9, Window: 24 * time.Hour, Query: "q"}
	repo.Create(context.Background(), slo)

	status, err := svc.EvaluateSLO(context.Background(), "s1")
	if err != nil {
		t.Fatalf("EvaluateSLO: %v", err)
	}
	if status.Compliance <= 0 {
		t.Fatalf("expected compliance > 0")
	}

// Basic expectation: recent status recorded and looks sane
if status.LastCalculated.IsZero() {
		t.Errorf("expected LastCalculated to be set")
}
}
