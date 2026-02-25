package grpcserver

import (
    "context"
    "net"
    "testing"

    obs "example.com/lma/observability-service/internal/gen/observability/v1/observability"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/test/bufconn"
)

func TestGRPCProto_GetTrace_OK(t *testing.T) {
    slo := &fakeSLOService2{}
    q := &fakeQueries2{}
    lis := bufconn.Listen(1<<20)
    srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier2{scopes: []string{"observability:read"}})))
    obs.RegisterObservabilityServiceServer(srv, NewProtoObservabilityServer(slo, &fakeAlertSvc{}, &fakeErrorSvc{}, q))
    go func(){ _ = srv.Serve(lis) }()
    dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
    cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
    if err != nil { t.Fatalf("dial: %v", err) }
    defer func(){ cc.Close(); srv.Stop(); lis.Close() }()
    ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization","Bearer t"))
    c := obs.NewObservabilityServiceClient(cc)
    out, err := c.GetTrace(ctx, &obs.GetTraceRequest{TraceId: "trace-1"})
    if err != nil { t.Fatalf("GetTrace: %v", err) }
    if out.GetTrace().GetTraceId() != "trace-1" { t.Fatalf("unexpected trace: %+v", out.GetTrace()) }
}

func TestGRPCProto_SearchLogs_OK(t *testing.T) {
    slo := &fakeSLOService2{}
    q := &fakeQueries2{}
    lis := bufconn.Listen(1<<20)
    srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier2{scopes: []string{"observability:read"}})))
    obs.RegisterObservabilityServiceServer(srv, NewProtoObservabilityServer(slo, &fakeAlertSvc{}, &fakeErrorSvc{}, q))
    go func(){ _ = srv.Serve(lis) }()
    dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
    cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
    if err != nil { t.Fatalf("dial: %v", err) }
    defer func(){ cc.Close(); srv.Stop(); lis.Close() }()
    ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization","Bearer t"))
    c := obs.NewObservabilityServiceClient(cc)
    out, err := c.SearchLogs(ctx, &obs.SearchLogsRequest{ProjectId: "p1", Q: "{app=\"api\"}", Limit: 1})
    if err != nil { t.Fatalf("SearchLogs: %v", err) }
    if len(out.GetLogs()) != 1 || out.GetLogs()[0].GetMessage() == "" { t.Fatalf("unexpected logs: %+v", out.GetLogs()) }
}

func TestGRPCProto_Golden_OK(t *testing.T) {
    slo := &fakeSLOService2{}
    q := &fakeQueries2{}
    lis := bufconn.Listen(1<<20)
    srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier2{scopes: []string{"observability:read"}})))
    obs.RegisterObservabilityServiceServer(srv, NewProtoObservabilityServer(slo, &fakeAlertSvc{}, &fakeErrorSvc{}, q))
    go func(){ _ = srv.Serve(lis) }()
    dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
    cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
    if err != nil { t.Fatalf("dial: %v", err) }
    defer func(){ cc.Close(); srv.Stop(); lis.Close() }()
    ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization","Bearer t"))
    c := obs.NewObservabilityServiceClient(cc)
    out, err := c.Golden(ctx, &obs.GoldenRequest{ProjectId: "p1"})
    if err != nil { t.Fatalf("Golden: %v", err) }
    m := out.GetDashboard().GetMetrics()
    if m["request_rate"] == 0 || m["p95_latency"] == 0 { t.Fatalf("unexpected dashboard: %+v", m) }
}

func TestGRPCProto_Queries_RequireReadScope(t *testing.T) {
    slo := &fakeSLOService2{}
    q := &fakeQueries2{}
    lis := bufconn.Listen(1<<20)
    srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier2{scopes: []string{"different:scope"}})))
    obs.RegisterObservabilityServiceServer(srv, NewProtoObservabilityServer(slo, &fakeAlertSvc{}, &fakeErrorSvc{}, q))
    go func(){ _ = srv.Serve(lis) }()
    dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
    cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
    if err != nil { t.Fatalf("dial: %v", err) }
    defer func(){ cc.Close(); srv.Stop(); lis.Close() }()
    ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization","Bearer t"))
    c := obs.NewObservabilityServiceClient(cc)
    if _, err := c.GetTrace(ctx, &obs.GetTraceRequest{TraceId: "t"}); status.Code(err) != codes.PermissionDenied { t.Fatalf("expected denied, got %v", err) }
    if _, err := c.SearchLogs(ctx, &obs.SearchLogsRequest{ProjectId: "p1", Q: "{}"}); status.Code(err) != codes.PermissionDenied { t.Fatalf("expected denied, got %v", err) }
    if _, err := c.Golden(ctx, &obs.GoldenRequest{ProjectId: "p1"}); status.Code(err) != codes.PermissionDenied { t.Fatalf("expected denied, got %v", err) }
}
