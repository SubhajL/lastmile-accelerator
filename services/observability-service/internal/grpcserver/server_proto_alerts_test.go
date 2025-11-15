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

type fakeAlertSvc struct{}

func (*fakeAlertSvc) CreateAlert(context.Context, *models.AlertRule) error { return nil }
func (*fakeAlertSvc) GetAlert(context.Context, string) (*models.AlertRule, error) {
    return &models.AlertRule{ID: "a1", SLOID: "s1", Threshold: 95, Channels: []models.AlertChannel{models.AlertChannelEmail}}, nil
}
func (*fakeAlertSvc) GetSLOAlerts(context.Context, string) ([]*models.AlertRule, error) {
    return []*models.AlertRule{{ID: "a1", SLOID: "s1", Threshold: 95, Channels: []models.AlertChannel{models.AlertChannelEmail}}}, nil
}
func (*fakeAlertSvc) UpdateAlert(context.Context, *models.AlertRule) error { return nil }
func (*fakeAlertSvc) DeleteAlert(context.Context, string) error { return nil }
func (*fakeAlertSvc) EvaluateAlerts(context.Context, string, *models.SLOStatus) error { return nil }
func (*fakeAlertSvc) NotifyAlert(context.Context, *models.AlertRule, *models.SLOStatus) error { return nil }
func (*fakeAlertSvc) GetAlertHistory(context.Context, string, int) ([]*models.AlertHistory, error) {
    return []*models.AlertHistory{{ID: "h1", AlertRuleID: "a1", SLOID: "s1", Compliance: 90, Threshold: 95, Notified: true, Timestamp: time.Now()}}, nil
}

func TestGRPCProto_Alerts_Create_Requires_Write_Scope(t *testing.T) {
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
    _, err = c.CreateAlert(ctx, &obs.CreateAlertRequest{Rule: &obs.AlertRule{SloId: "s1", Threshold: 95}})
    if status.Code(err) != codes.PermissionDenied {
        t.Fatalf("expected permission denied, got %v", err)
    }
}

func TestGRPCProto_Alerts_CRUD_Lifecycle_Succeeds(t *testing.T) {
    slo := &fakeSLOService2{}
    q := &fakeQueries2{}
    lis := bufconn.Listen(1 << 20)
    srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier2{scopes: []string{"observability:read","observability:write"}})))
    obs.RegisterObservabilityServiceServer(srv, NewProtoObservabilityServer(slo, &fakeAlertSvc{}, &fakeErrorSvc{}, q))
    go func(){ _ = srv.Serve(lis) }()
    dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
    cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
    if err != nil { t.Fatalf("dial: %v", err) }
    defer func(){ cc.Close(); srv.Stop(); lis.Close() }()
    ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer t"))
    c := obs.NewObservabilityServiceClient(cc)

    // Create
    if _, err := c.CreateAlert(ctx, &obs.CreateAlertRequest{Rule: &obs.AlertRule{Id:"a1", SloId: "s1", Threshold: 95}}); err != nil { t.Fatalf("create: %v", err) }
    // Get
    if _, err := c.GetAlert(ctx, &obs.GetAlertRequest{Id: "a1"}); err != nil { t.Fatalf("get: %v", err) }
    // List
    list, err := c.ListAlertsBySLO(ctx, &obs.ListAlertsBySLORequest{SloId: "s1"})
    if err != nil || len(list.GetRules()) == 0 { t.Fatalf("list: %v", err) }
    // Update
    if _, err := c.UpdateAlert(ctx, &obs.UpdateAlertRequest{Rule: &obs.AlertRule{Id:"a1", SloId: "s1", Threshold: 90}}); err != nil { t.Fatalf("update: %v", err) }
    // History
    if _, err := c.GetAlertHistory(ctx, &obs.GetAlertHistoryRequest{Id: "a1"}); err != nil { t.Fatalf("history: %v", err) }
    // Delete
    if _, err := c.DeleteAlert(ctx, &obs.DeleteAlertRequest{Id: "a1"}); err != nil { t.Fatalf("delete: %v", err) }
}
