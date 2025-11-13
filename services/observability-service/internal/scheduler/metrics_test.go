package scheduler

import (
    "testing"
    "time"
)

type testRecorder struct{
    evaluated int
    errors    int
    dCount    int
    lastRun   time.Time
}

func (t *testRecorder) IncEvaluated(n int)            { t.evaluated += n }
func (t *testRecorder) IncErrors(n int)               { t.errors += n }
func (t *testRecorder) ObserveDuration(d time.Duration) { if d > 0 { t.dCount++ } }
func (t *testRecorder) SetLastRun(ts time.Time)       { t.lastRun = ts }

func TestUseMetricsRecorder_SetsGlobal(t *testing.T) {
    // Restore noop after test
    t.Cleanup(func(){ UseMetricsRecorder(nil) })

    r := &testRecorder{}
    UseMetricsRecorder(r)

    // Exercise through the package accessor
    getRecorder().IncEvaluated(2)
    getRecorder().IncErrors(1)
    getRecorder().ObserveDuration(10 * time.Millisecond)
    getRecorder().SetLastRun(time.Now())

    if r.evaluated != 2 { t.Fatalf("evaluated counter not updated: %d", r.evaluated) }
    if r.errors != 1 { t.Fatalf("errors counter not updated: %d", r.errors) }
    if r.dCount == 0 { t.Fatalf("duration observation not recorded") }
    if r.lastRun.IsZero() { t.Fatalf("lastRun not set") }
}
