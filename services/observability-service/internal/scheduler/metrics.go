package scheduler

import (
    "sync"
    "time"
)

// MetricsRecorder captures minimal metrics for the SLO scheduler.
type MetricsRecorder interface {
    IncEvaluated(n int)
    IncErrors(n int)
    ObserveDuration(d time.Duration)
    SetLastRun(t time.Time)
}

var (
    recMu    sync.RWMutex
    recorder MetricsRecorder = noopRecorder{}
)

// UseMetricsRecorder sets the package-level metrics recorder.
// Passing nil installs a no-op recorder.
func UseMetricsRecorder(r MetricsRecorder) {
    recMu.Lock()
    defer recMu.Unlock()
    if r == nil { recorder = noopRecorder{} } else { recorder = r }
}

func getRecorder() MetricsRecorder {
    recMu.RLock()
    defer recMu.RUnlock()
    return recorder
}

type noopRecorder struct{}

func (noopRecorder) IncEvaluated(int)            {}
func (noopRecorder) IncErrors(int)               {}
func (noopRecorder) ObserveDuration(time.Duration) {}
func (noopRecorder) SetLastRun(time.Time)        {}
