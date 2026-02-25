package grpcserver

import (
    "context"
    "net"
    "testing"
    "time"

    obs "example.com/lma/observability-service/internal/gen/observability/v1/observability"
    "example.com/lma/observability-service/internal/models"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/test/bufconn"
    "google.golang.org/protobuf/types/known/durationpb"
)

type memSLO struct{ byID map[string]*models.SLO }

func (m *memSLO) CreateSLO(_ context.Context, s *models.SLO) error { if m.byID==nil { m.byID = map[string]*models.SLO{} }; m.byID[s.ID] = s; return nil }
func (m *memSLO) GetSLO(_ context.Context, id string) (*models.SLO, error) { if s, ok := m.byID[id]; ok { return s, nil }; return nil, nil }
func (m *memSLO) ListProjectSLOs(_ context.Context, projectID string) ([]*models.SLO, error) { var out []*models.SLO; for _, s := range m.byID { if s.ProjectID==projectID { out = append(out, s) } }; return out, nil }
func (m *memSLO) UpdateSLO(_ context.Context, s *models.SLO) error { m.byID[s.ID] = s; return nil }
func (m *memSLO) DeleteSLO(_ context.Context, id string) error { delete(m.byID, id); return nil }
func (m *memSLO) EvaluateSLO(context.Context, string) (*models.SLOStatus, error) { return nil, nil }
func (m *memSLO) GetSLOStatus(context.Context, string) (*models.SLOStatus, error) { return nil, nil }
func (m *memSLO) GetSLOHistory(context.Context, string, time.Time, time.Time) ([]*models.SLOHistory, error) { return nil, nil }

func TestGRPCProto_SLO_CRUD_OK(t *testing.T) {
    slo := &memSLO{}
    q := &fakeQueries2{}
    lis := bufconn.Listen(1<<20)
    srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier2{scopes: []string{"observability:write","observability:read"}})))
    obs.RegisterObservabilityServiceServer(srv, NewProtoObservabilityServer(slo, &fakeAlertSvc{}, &fakeErrorSvc{}, q))
    go func(){ _ = srv.Serve(lis) }()
    dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
    cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer))
    if err != nil { t.Fatalf("dial: %v", err) }
    defer func(){ cc.Close(); srv.Stop(); lis.Close() }()
    ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization","Bearer t"))

    c := obs.NewObservabilityServiceClient(cc)
    // Create
    _, err = c.CreateSLO(ctx, &obs.CreateSLORequest{Slo: &obs.SLO{Id:"s1", ProjectId:"p1", ServiceName:"svc", Type: obs.SLOType_AVAILABILITY, Target: 99.9, Window: durationpb.New(5*time.Minute), Query:"up", Status:"active"}})
    if err != nil { t.Fatalf("CreateSLO: %v", err) }
    // Get
    got, err := c.GetSLO(ctx, &obs.GetSLORequest{Id:"s1"})
    if err != nil { t.Fatalf("GetSLO: %v", err) }
    if got.GetSlo().GetServiceName() != "svc" || got.GetSlo().GetTarget() < 99 {
        t.Fatalf("bad slo: %+v", got.GetSlo())
    }
    // Update
    _, err = c.UpdateSLO(ctx, &obs.UpdateSLORequest{Slo: &obs.SLO{Id:"s1", ProjectId:"p1", ServiceName:"svc2", Type: obs.SLOType_AVAILABILITY, Target: 99.9, Window: durationpb.New(5*time.Minute), Query:"up", Status:"active"}})
    if err != nil { t.Fatalf("UpdateSLO: %v", err) }
    got2, _ := c.GetSLO(ctx, &obs.GetSLORequest{Id:"s1"})
    if got2.GetSlo().GetServiceName() != "svc2" { t.Fatalf("update not applied: %+v", got2.GetSlo()) }
    // List
    list, err := c.ListSLOsByProject(ctx, &obs.ListSLOsByProjectRequest{ProjectId:"p1"})
    if err != nil || len(list.GetSlos()) != 1 { t.Fatalf("list: %v %v", err, len(list.GetSlos())) }
    // Delete
    if _, err := c.DeleteSLO(ctx, &obs.DeleteSLORequest{Id:"s1"}); err != nil { t.Fatalf("delete: %v", err) }
}
