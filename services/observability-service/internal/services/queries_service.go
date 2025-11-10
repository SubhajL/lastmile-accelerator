package services

import (
	"context"
	"fmt"
	"time"

	"example.com/lma/observability-service/internal/logs"
	"example.com/lma/observability-service/internal/traces"
)

type TracesClient interface{
	SearchTraceQL(ctx context.Context, query string, start, end time.Time, limit int) ([]traces.TraceSummary, error)
	GetTrace(ctx context.Context, traceID string) (*traces.TraceDetail, error)
}

type LogsClient interface{
	QueryRange(ctx context.Context, q string, start, end time.Time, limit int, direction string) ([]logs.LogEntry, error)
}

type PromQuerier interface{
	Query(ctx context.Context, promQL string, window time.Duration) (float64, error)
}

type ObservabilityQueries struct{
	traces TracesClient
	logs   LogsClient
	prom   PromQuerier
}

func NewObservabilityQueries(t TracesClient, l LogsClient, p PromQuerier) *ObservabilityQueries {
	return &ObservabilityQueries{traces: t, logs: l, prom: p}
}

// BuildTraceQL constructs a simple TraceQL from filters.
func BuildTraceQL(service, operation string, minDur, maxDur time.Duration, errorOnly bool, userQuery string) string {
	if userQuery != "" { return userQuery }
	q := "{"
	first := true
	add := func(s string){ if !first { q += " && " }; q += s; first = false }
	if service != "" { add(fmt.Sprintf(`service.name == "%s"`, service)) }
	if operation != "" { add(fmt.Sprintf(`span.name == "%s"`, operation)) }
	if minDur > 0 { add(fmt.Sprintf(`duration >= %dms`, minDur.Milliseconds())) }
	if maxDur > 0 { add(fmt.Sprintf(`duration <= %dms`, maxDur.Milliseconds())) }
	if errorOnly { add(`status == error`) }
	q += "}"
	return q
}

func (o *ObservabilityQueries) SearchTraces(ctx context.Context, service, operation string, minDur, maxDur time.Duration, errorOnly bool, userQuery string, start, end time.Time, limit int) ([]traces.TraceSummary, error) {
	q := BuildTraceQL(service, operation, minDur, maxDur, errorOnly, userQuery)
	return o.traces.SearchTraceQL(ctx, q, start, end, limit)
}

func (o *ObservabilityQueries) GetTrace(ctx context.Context, id string) (*traces.TraceDetail, error) {
	return o.traces.GetTrace(ctx, id)
}

func (o *ObservabilityQueries) SearchLogs(ctx context.Context, logQL string, start, end time.Time, limit int, direction string) ([]logs.LogEntry, error) {
	return o.logs.QueryRange(ctx, logQL, start, end, limit, direction)
}

type GoldenSignals struct {
	RequestRate float64 `json:"request_rate"`
	ErrorRate   float64 `json:"error_rate"`
	P95Latency  float64 `json:"p95_latency"`
}

func (o *ObservabilityQueries) Golden(ctx context.Context, service string, window time.Duration) (*GoldenSignals, error) {
	if service == "" { return nil, fmt.Errorf("service required") }
	w := fmt.Sprintf("%dm", int(window.Minutes()))
	rpsQ := fmt.Sprintf(`sum(rate(http_requests_total{service="%s"}[%s]))`, service, w)
	errQ := fmt.Sprintf(`sum(rate(http_requests_total{service="%s",status=~"5.."}[%s])) / clamp_min(sum(rate(http_requests_total{service="%s"}[%s])), 1)`, service, w, service, w)
	latQ := fmt.Sprintf(`histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{service="%s"}[%s])) by (le))`, service, w)
	rps, err := o.prom.Query(ctx, rpsQ, window); if err != nil { return nil, err }
	errRate, err := o.prom.Query(ctx, errQ, window); if err != nil { return nil, err }
	p95, err := o.prom.Query(ctx, latQ, window); if err != nil { return nil, err }
	return &GoldenSignals{RequestRate: rps, ErrorRate: errRate, P95Latency: p95}, nil
}