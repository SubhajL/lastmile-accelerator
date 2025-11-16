package grpcserver

import (
    "context"
    "net"
    "testing"
    "time"

    obs "example.com/lma/observability-service/internal/gen/observability/v1/observability"
    "example.com/lma/observability-service/internal/middleware"
    "example.com/lma/observability-service/internal/models"
    "example.com/lma/observability-service/internal/services"
    "example.com/lma/observability-service/internal/logs"
    "example.com/lma/observability-service/internal/traces"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/test/bufconn"
    "google.golang.org/protobuf/types/known/timestamppb"
)

type fakeVerifier2 struct{ scopes []string; err error }

func (f *fakeVerifier2) Verify(ctx context.Context, token string) (*middleware.Claims, error) {
    if f.err != nil { return nil, f.err }
    return &middleware.Claims{Subject: "u", Scopes: f.scopes}, nil
}

type fakeSLOService2 struct{ resp *models.SLOStatus }

func (f *fakeSLOService2) CreateSLO(context.Context, *models.SLO) error { return nil }
func (f *fakeSLOService2) GetSLO(context.Context, string) (*models.SLO, error) { return nil, nil }
func (f *fakeSLOService2) ListProjectSLOs(context.Context, string) ([]*models.SLO, error) { return nil, nil }
func (f *fakeSLOService2) UpdateSLO(context.Context, *models.SLO) error { return nil }
func (f *fakeSLOService2) DeleteSLO(context.Context, string) error { return nil }
func (f *fakeSLOService2) EvaluateSLO(context.Context, string) (*models.SLOStatus, error) { return f.resp, nil }
func (f *fakeSLOService2) GetSLOStatus(context.Context, string) (*models.SLOStatus, error) { return f.resp, nil }
func (f *fakeSLOService2) GetSLOHistory(context.Context, string, time.Time, time.Time) ([]*models.SLOHistory, error) { return nil, nil }

type fakeQueries2 struct{}

func (f *fakeQueries2) SearchTraces(ctx context.Context, service, operation string, minDur, maxDur time.Duration, errorOnly bool, userQuery string, start, end time.Time, limit int) ([]traces.TraceSummary, error) {
    return []traces.TraceSummary{{TraceID: "t1", Service: "svc", Operation: "op", DurationMs: 10, Error: false}}, nil
}
func (f *fakeQueries2) GetTrace(ctx context.Context, id string) (*traces.TraceDetail, error) { return &traces.TraceDetail{TraceID: id}, nil }
func (f *fakeQueries2) SearchLogs(ctx context.Context, q string, start, end time.Time, limit int, direction string) ([]logs.LogEntry, error) {
    return []logs.LogEntry{{Timestamp: time.Now(), Line: "hello", Labels: map[string]string{"lvl":"info"}}}, nil
}
func (f *fakeQueries2) Golden(ctx context.Context, service string, window time.Duration) (*services.GoldenSignals, error) { return &services.GoldenSignals{RequestRate: 1, ErrorRate: 0, P95Latency: 0.1}, nil }

// fake alert/error services are defined in other *_test.go files in this package

func dialBufConnProto(t *testing.T, slo *fakeSLOService2, q *fakeQueries2, scopes []string) (context.Context, *grpc.ClientConn, func()) {
    t.Helper()
    lis := bufconn.Listen(1 << 20)
    srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier2{scopes: scopes})))
    obs.RegisterObservabilityServiceServer(srv, NewProtoObservabilityServer(slo, &fakeAlertSvc{}, &fakeErrorSvc{}, q))
    go func(){ _ = srv.Serve(lis) }()
    dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
    cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
    if err != nil { t.Fatalf("dial: %v", err) }
    ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer token"))
    cleanup := func(){ cc.Close(); srv.Stop(); lis.Close() }
    return ctx, cc, cleanup
}

func TestGRPCProto_GetSLOStatus_OK(t *testing.T) {
    slo := &fakeSLOService2{resp: &models.SLOStatus{SLOID: "s1", Compliance: 99, BurnRate: 0.1, RemainingBudget: 1.0}}
    q := &fakeQueries2{}
    ctx, cc, done := dialBufConnProto(t, slo, q, []string{"observability:read"})
    defer done()
    c := obs.NewObservabilityServiceClient(cc)
    out, err := c.GetSLOStatus(ctx, &obs.GetSLOStatusRequest{Id: "s1"})
    if err != nil { t.Fatalf("GetSLOStatus: %v", err) }
    if out.GetStatus().GetSloId() != "s1" || out.GetStatus().GetCompliance() <= 0 {
        t.Fatalf("bad status: %+v", out)
    }
}

func TestGRPCProto_SearchTraces_OK(t *testing.T) {
    slo := &fakeSLOService2{}
    q := &fakeQueries2{}
    ctx, cc, done := dialBufConnProto(t, slo, q, []string{"observability:read"})
    defer done()
    c := obs.NewObservabilityServiceClient(cc)
    now := time.Now()
    out, err := c.SearchTraces(ctx, &obs.SearchTracesRequest{Q: "{service.name == \"svc\"}", Range: &obs.TimeRange{From: timestamppb.New(now.Add(-time.Minute)), To: timestamppb.New(now)}, Limit: 1})
    if err != nil { t.Fatalf("SearchTraces: %v", err) }
    if len(out.GetTraces()) != 1 || out.GetTraces()[0].GetTraceId() != "t1" || out.GetTraces()[0].GetServiceName() != "svc" {
        t.Fatalf("bad traces: %+v", out.GetTraces())
    }
}

func TestGRPCProto_Auth_WrongScope_PermissionDenied(t *testing.T) {
    slo := &fakeSLOService2{resp: &models.SLOStatus{SLOID: "s1"}}
    q := &fakeQueries2{}
    ctx, cc, done := dialBufConnProto(t, slo, q, []string{"different:scope"})
    defer done()
    c := obs.NewObservabilityServiceClient(cc)
    _, err := c.GetSLOStatus(ctx, &obs.GetSLOStatusRequest{Id: "s1"})
    if status.Code(err) != codes.PermissionDenied {
        t.Fatalf("expected permission denied, got %v", err)
    }
}

func TestGRPCProto_Auth_MissingToken_Unauthenticated(t *testing.T) {
    slo := &fakeSLOService2{resp: &models.SLOStatus{SLOID: "s1"}}
    q := &fakeQueries2{}
    // start server with auth interceptor but call without metadata
    lis := bufconn.Listen(1 << 20)
    srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier2{scopes: []string{"observability:read"}})))
    obs.RegisterObservabilityServiceServer(srv, NewProtoObservabilityServer(slo, &fakeAlertSvc{}, &fakeErrorSvc{}, q))
    go func(){ _ = srv.Serve(lis) }()
    dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
    cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
    if err != nil { t.Fatalf("dial: %v", err) }
    defer func(){ cc.Close(); srv.Stop(); lis.Close() }()
    c := obs.NewObservabilityServiceClient(cc)
    _, err = c.GetSLOStatus(context.Background(), &obs.GetSLOStatusRequest{Id: "s1"})
    if status.Code(err) != codes.Unauthenticated {
        t.Fatalf("expected unauthenticated, got %v", err)
    }
}
