package scheduler

import (
    "context"
    "errors"
    "testing"

    "example.com/lma/observability-service/internal/models"
)

// Reuse fakeRepo and fakeSvc patterns from existing tests in this package.

func TestSLOEvaluator_RecordsMetrics_OnRunOnce(t *testing.T) {
    t.Cleanup(func(){ UseMetricsRecorder(nil) })

    repo := &fakeRepo{list: []*models.SLO{{ID: "a"}, {ID: "b"}}}
    svc := &fakeSvc{errOn: map[string]error{"b": errors.New("boom")}}
    e := NewSLOEvaluator(repo, svc)

    rec := &testRecorder{}
    UseMetricsRecorder(rec)

    _ = e.runOnce(context.Background())

    if rec.evaluated != 2 { t.Fatalf("want evaluated=2 got=%d", rec.evaluated) }
    if rec.errors != 1 { t.Fatalf("want errors=1 got=%d", rec.errors) }
    if rec.dCount == 0 { t.Fatalf("duration not observed") }
    if rec.lastRun.IsZero() { t.Fatalf("last run not set") }
}
