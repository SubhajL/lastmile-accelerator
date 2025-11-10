package scheduler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/repository"
	"example.com/lma/observability-service/internal/services"
)

type fakeRepo struct{ list []*models.SLO; err error }

func (f *fakeRepo) Create(context.Context, *models.SLO) error { return nil }
func (f *fakeRepo) GetByID(context.Context, string) (*models.SLO, error) { return nil, nil }
func (f *fakeRepo) ListByProject(context.Context, string) ([]*models.SLO, error) { return nil, nil }
func (f *fakeRepo) ListAll(context.Context) ([]*models.SLO, error) { return f.list, f.err }
func (f *fakeRepo) Update(context.Context, *models.SLO) error { return nil }
func (f *fakeRepo) Delete(context.Context, string) error { return nil }
func (f *fakeRepo) SaveStatus(context.Context, *models.SLOStatus) error { return nil }
func (f *fakeRepo) GetStatus(context.Context, string) (*models.SLOStatus, error) { return nil, nil }
func (f *fakeRepo) GetHistory(context.Context, string, time.Time, time.Time) ([]*models.SLOHistory, error) {
	return nil, nil
}

type fakeSvc struct{
	calls []string
	sleep time.Duration
	errOn map[string]error
	mu sync.Mutex
}

func (s *fakeSvc) CreateSLO(context.Context, *models.SLO) error { return nil }
func (s *fakeSvc) GetSLO(context.Context, string) (*models.SLO, error) { return nil, nil }
func (s *fakeSvc) ListProjectSLOs(context.Context, string) ([]*models.SLO, error) { return nil, nil }
func (s *fakeSvc) UpdateSLO(context.Context, *models.SLO) error { return nil }
func (s *fakeSvc) DeleteSLO(context.Context, string) error { return nil }
func (s *fakeSvc) EvaluateSLO(ctx context.Context, id string) (*models.SLOStatus, error) {
	t := time.NewTimer(s.sleep)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-t.C:
	}
	s.mu.Lock(); s.calls = append(s.calls, id); s.mu.Unlock()
	if s.errOn != nil {
		if err := s.errOn[id]; err != nil { return nil, err }
	}
	return &models.SLOStatus{SLOID: id, LastCalculated: time.Now()}, nil
}
func (s *fakeSvc) GetSLOStatus(context.Context, string) (*models.SLOStatus, error) { return nil, nil }
func (s *fakeSvc) GetSLOHistory(context.Context, string, time.Time, time.Time) ([]*models.SLOHistory, error) {
	return nil, nil
}

var _ repository.SLORepository = (*fakeRepo)(nil)
var _ services.SLOService = (*fakeSvc)(nil)

func TestSLOEvaluator_RunOnce_EvaluatesAll(t *testing.T) {
	repo := &fakeRepo{list: []*models.SLO{{ID: "s1"}, {ID: "s2"}, {ID: "s3"}}}
	svc := &fakeSvc{}
	e := NewSLOEvaluator(repo, svc)
	if err := e.runOnce(context.Background()); err != nil {
		t.Fatalf("runOnce: %v", err)
	}
	if len(svc.calls) != 3 { t.Fatalf("want 3 calls, got %d", len(svc.calls)) }
}

func TestSLOEvaluator_RunOnce_ErrorAggregation(t *testing.T) {
	repo := &fakeRepo{list: []*models.SLO{{ID: "a"}, {ID: "b"}}}
	svc := &fakeSvc{errOn: map[string]error{"b": errors.New("boom")}}
	e := NewSLOEvaluator(repo, svc)
	err := e.runOnce(context.Background())
	if err == nil { t.Fatal("expected error") }
	if len(svc.calls) != 2 { t.Fatalf("expected both evaluated, got %d", len(svc.calls)) }
}

func TestSLOEvaluator_Run_CancellationStopsLoop(t *testing.T) {
	repo := &fakeRepo{list: []*models.SLO{{ID: "x"}}}
	svc := &fakeSvc{sleep: 5 * time.Millisecond}
	e := NewSLOEvaluator(repo, svc)
	ctx, cancel := context.WithCancel(context.Background())
	go func(){ _ = e.Run(ctx, 5*time.Millisecond, 2) }()
	time.Sleep(15 * time.Millisecond)
	cancel()
	time.Sleep(5 * time.Millisecond)
	if len(svc.calls) == 0 {
		t.Fatal("expected at least one evaluation call before cancel")
	}
}

func TestSLOEvaluator_RunBatch_RespectsWorkers(t *testing.T) {
	repo := &fakeRepo{}
	svc := &fakeSvc{sleep: 50 * time.Millisecond}
	e := NewSLOEvaluator(repo, svc)
	ids := []string{"1","2","3","4"}

	start1 := time.Now()
	_ = e.runBatch(context.Background(), ids, 1)
	d1 := time.Since(start1)

	start2 := time.Now()
	_ = e.runBatch(context.Background(), ids, 2)
	d2 := time.Since(start2)

	if !(d2 < d1) {
		t.Fatalf("expected faster with 2 workers: d1=%v d2=%v", d1, d2)
	}
}
