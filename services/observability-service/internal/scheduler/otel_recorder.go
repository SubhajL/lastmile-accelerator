package scheduler

import (
    "context"
    "sync/atomic"
    "time"

    "go.opentelemetry.io/otel/metric"
)

type otelRecorder struct {
    eval     metric.Int64Counter
    errs     metric.Int64Counter
    duration metric.Float64Histogram
    lastRun  atomic.Int64 // unix seconds
    _reg     metric.Registration
}

// NewOTelRecorder constructs a MetricsRecorder backed by OTel instruments.
func NewOTelRecorder(m meter) MetricsRecorder { return newOTelRecorder(m) }

// meter abstracts metric.Meter for testability.
type meter interface {
    Int64Counter(string, ...metric.Int64CounterOption) (metric.Int64Counter, error)
    Int64ObservableGauge(string, ...metric.Int64ObservableGaugeOption) (metric.Int64ObservableGauge, error)
    Float64Histogram(string, ...metric.Float64HistogramOption) (metric.Float64Histogram, error)
    RegisterCallback(metric.Callback, ...metric.Observable) (metric.Registration, error)
}

func newOTelRecorder(m meter) MetricsRecorder {
    eval, _ := m.Int64Counter("scheduler.slo_evaluated_total",
        metric.WithUnit("1"),
        metric.WithDescription("Total SLOs evaluated by scheduler"),
    )
    errs, _ := m.Int64Counter("scheduler.slo_errors_total",
        metric.WithUnit("1"),
        metric.WithDescription("Total SLO evaluation errors"),
    )
    dur, _ := m.Float64Histogram("scheduler.evaluation.duration_ms",
        metric.WithUnit("ms"),
        metric.WithDescription("Evaluation duration in milliseconds"),
    )
    gauge, _ := m.Int64ObservableGauge("scheduler.last_run_unixtime_s",
        metric.WithUnit("s"),
        metric.WithDescription("Last evaluation run end time (unix seconds)"),
    )

    r := &otelRecorder{eval: eval, errs: errs, duration: dur}
    cb := metric.Callback(func(ctx context.Context, o metric.Observer) error {
        o.ObserveInt64(gauge, r.lastRun.Load())
        return nil
    })
    reg, _ := m.RegisterCallback(cb, gauge)
    r._reg = reg
    return r
}

func (r *otelRecorder) IncEvaluated(n int) {
    if n <= 0 { return }
    r.eval.Add(context.Background(), int64(n))
}

func (r *otelRecorder) IncErrors(n int) {
    if n <= 0 { return }
    r.errs.Add(context.Background(), int64(n))
}

func (r *otelRecorder) ObserveDuration(d time.Duration) {
    if d <= 0 { return }
    r.duration.Record(context.Background(), float64(d.Milliseconds()))
}

func (r *otelRecorder) SetLastRun(t time.Time) {
    if t.IsZero() { return }
    r.lastRun.Store(t.Unix())
}
