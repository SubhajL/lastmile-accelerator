package grpcserver

import (
    "context"
    "testing"
    "time"

    dto "github.com/prometheus/client_model/go"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/testutil"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func TestMetricsUnaryInterceptor_RecordsSuccessCountAndLatency(t *testing.T) {
    intr := MetricsUnaryInterceptor()
    method := "/test.Metrics/Success"
    info := &grpc.UnaryServerInfo{FullMethod: method}
    h := func(ctx context.Context, req interface{}) (interface{}, error) { time.Sleep(1 * time.Millisecond); return nil, nil }
    if _, err := intr(context.Background(), nil, info, h); err != nil { t.Fatalf("unexpected: %v", err) }

    // Counter incremented for OK
    got := testutil.ToFloat64(grpcRequestsTotal.WithLabelValues(method, codes.OK.String()))
    if got < 1 { t.Fatalf("expected counter >=1, got %v", got) }
    // Histogram count increased
    if c := histogramCountForLabels(t, method, codes.OK.String()); c < 1 { t.Fatalf("expected hist count >=1, got %d", c) }
}

func TestMetricsUnaryInterceptor_RecordsErrorCodeCount(t *testing.T) {
    intr := MetricsUnaryInterceptor()
    method := "/test.Metrics/Error"
    info := &grpc.UnaryServerInfo{FullMethod: method}
    h := func(ctx context.Context, req interface{}) (interface{}, error) { return nil, status.Error(codes.NotFound, "nope") }
    if _, err := intr(context.Background(), nil, info, h); status.Code(err) != codes.NotFound { t.Fatalf("unexpected: %v", err) }

    got := testutil.ToFloat64(grpcRequestsTotal.WithLabelValues(method, codes.NotFound.String()))
    if got < 1 { t.Fatalf("expected counter >=1, got %v", got) }
}

func TestMetricsUnaryInterceptor_AccumulatesAcrossMultipleCalls(t *testing.T) {
    intr := MetricsUnaryInterceptor()
    method := "/test.Metrics/Multi"
    info := &grpc.UnaryServerInfo{FullMethod: method}
    h := func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil }
    for i := 0; i < 3; i++ { _, _ = intr(context.Background(), nil, info, h) }
    got := testutil.ToFloat64(grpcRequestsTotal.WithLabelValues(method, codes.OK.String()))
    if got < 3 { t.Fatalf("expected counter >=3, got %v", got) }
}

// histogramCountForLabels finds the SampleCount for the histogram with given labels.
func histogramCountForLabels(t *testing.T, method, code string) uint64 {
    t.Helper()
    mfs, err := prometheus.DefaultGatherer.Gather()
    if err != nil { t.Fatalf("gather: %v", err) }
    // Explicit name match (safe against Desc changes):
    var target *dto.MetricFamily
    for _, mf := range mfs { if mf.GetName() == "grpc_request_duration_seconds" { target = mf; break } }
    if target == nil { t.Fatalf("histogram family not found") }
    for _, m := range target.Metric {
        if matchLabels(m, method, code) {
            if m.GetHistogram() == nil { t.Fatalf("expected histogram") }
            return m.GetHistogram().GetSampleCount()
        }
    }
    return 0
}

func matchLabels(m *dto.Metric, method, code string) bool {
    var gotMethod, gotCode string
    for _, lp := range m.Label {
        if lp.GetName() == "method" { gotMethod = lp.GetValue() }
        if lp.GetName() == "code" { gotCode = lp.GetValue() }
    }
    return gotMethod == method && gotCode == code
}
