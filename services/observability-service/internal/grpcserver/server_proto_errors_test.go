package grpcserver

import (
    "context"
    "net"
    "testing"
    "time"

    obs "example.com/lma/observability-service/internal/gen/observability/v1/observability"
    "example.com/lma/observability-service/internal/models"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/test/bufconn"
)

type fakeErrorSvc struct{}

func (*fakeErrorSvc) Ingest(context.Context, string, string, string, string, string, map[string]interface{}) (string, error) { return "g1", nil }
func (*fakeErrorSvc) ListGroups(context.Context, string, models.ErrorGroupFilter) ([]models.ErrorGroup, error) {
    return []models.ErrorGroup{{ID: "g1", ProjectID: "p1", Fingerprint: "fp", Title: "t", Status: "open", FirstSeen: time.Now().Add(-time.Hour).Format(time.RFC3339), LastSeen: time.Now().Format(time.RFC3339), Occurrences: 1}}, nil
}
func (*fakeErrorSvc) GetGroup(context.Context, string) (*models.ErrorGroup, error) { return &models.ErrorGroup{ID: "g1"}, nil }
func (*fakeErrorSvc) ListGroupEvents(context.Context, string, int, int) ([]models.ErrorEvent, error) {
    return []models.ErrorEvent{{ID: "e1", GroupID: "g1", ProjectID: "p1", TS: time.Now().Format(time.RFC3339), Message: "m", Stack: "s", Metadata: map[string]interface{}{"k":"v"}}}, nil
}
func (*fakeErrorSvc) ResolveGroup(context.Context, string) error { return nil }

func TestGRPCProto_Errors_Ingest_Requires_Ingest_Scope(t *testing.T) {
    slo := &fakeSLOService2{}
    q := &fakeQueries2{}
    lis := bufconn.Listen(1 << 20)
    srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier2{scopes: []string{"observability:read"}})))
    obs.RegisterObservabilityServiceServer(srv, NewProtoObservabilityServer(slo, &fakeAlertSvc{}, &fakeErrorSvc{}, q))
    go func(){ _ = srv.Serve(lis) }()
    dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
    cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
    if err != nil { t.Fatalf("dial: %v", err) }
    defer func(){ cc.Close(); srv.Stop(); lis.Close() }()
    c := obs.NewObservabilityServiceClient(cc)
    ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer t"))
    _, err = c.IngestError(ctx, &obs.IngestErrorRequest{ProjectId: "p1", Event: &obs.ErrorEvent{Message: "boom"}})
    if status.Code(err) != codes.PermissionDenied { t.Fatalf("expected denied, got %v", err) }
}

func TestGRPCProto_Errors_Flows_Succeed_With_Scopes(t *testing.T) {
    slo := &fakeSLOService2{}
    q := &fakeQueries2{}
    lis := bufconn.Listen(1 << 20)
    srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier2{scopes: []string{"observability:read","observability:write","observability:ingest"}})))
    obs.RegisterObservabilityServiceServer(srv, NewProtoObservabilityServer(slo, &fakeAlertSvc{}, &fakeErrorSvc{}, q))
    go func(){ _ = srv.Serve(lis) }()
    dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
    cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
    if err != nil { t.Fatalf("dial: %v", err) }
    defer func(){ cc.Close(); srv.Stop(); lis.Close() }()
    ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer t"))
    c := obs.NewObservabilityServiceClient(cc)
    if _, err := c.IngestError(ctx, &obs.IngestErrorRequest{ProjectId: "p1", Event: &obs.ErrorEvent{Message: "m"}}); err != nil { t.Fatalf("ingest: %v", err) }
    if _, err := c.ListErrorGroups(ctx, &obs.ListErrorGroupsRequest{ProjectId: "p1"}); err != nil { t.Fatalf("list: %v", err) }
    if _, err := c.GetErrorGroup(ctx, &obs.GetErrorGroupRequest{GroupId: "g1"}); err != nil { t.Fatalf("get: %v", err) }
    if _, err := c.ListGroupEvents(ctx, &obs.ListGroupEventsRequest{GroupId: "g1", Limit: 10}); err != nil { t.Fatalf("events: %v", err) }
    if _, err := c.ResolveGroup(ctx, &obs.ResolveGroupRequest{GroupId: "g1"}); err != nil { t.Fatalf("resolve: %v", err) }
}
