package grpcserver

import (
    "context"
    "time"

    "example.com/lma/observability-service/internal/logs"
    "example.com/lma/observability-service/internal/services"
    "example.com/lma/observability-service/internal/traces"
)

// TraceQuery defines the query capabilities used by the gRPC proto server
// to serve trace, log, and golden dashboard requests.
type TraceQuery interface {
    SearchTraces(ctx context.Context, service, operation string, minDur, maxDur time.Duration, errorOnly bool, userQuery string, start, end time.Time, limit int) ([]traces.TraceSummary, error)
    GetTrace(ctx context.Context, id string) (*traces.TraceDetail, error)
    SearchLogs(ctx context.Context, q string, start, end time.Time, limit int, direction string) ([]logs.LogEntry, error)
    Golden(ctx context.Context, service string, window time.Duration) (*services.GoldenSignals, error)
}
