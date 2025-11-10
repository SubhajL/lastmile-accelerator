package grpcserver

import (
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"example.com/lma/observability-service/internal/logs"
	"example.com/lma/observability-service/internal/middleware"
	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/services"
	"example.com/lma/observability-service/internal/traces"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

type fakeVerifier struct{ scopes []string; err error }

func (f *fakeVerifier) Verify(ctx context.Context, token string) (*middleware.Claims, error) {
	if f.err != nil { return nil, f.err }
	return &middleware.Claims{Subject: "u", Scopes: f.scopes}, nil
}

type fakeSLOService struct{ resp *models.SLOStatus }

func (f *fakeSLOService) CreateSLO(context.Context, *models.SLO) error { return nil }
func (f *fakeSLOService) GetSLO(context.Context, string) (*models.SLO, error) { return nil, nil }
func (f *fakeSLOService) ListProjectSLOs(context.Context, string) ([]*models.SLO, error) { return nil, nil }
func (f *fakeSLOService) UpdateSLO(context.Context, *models.SLO) error { return nil }
func (f *fakeSLOService) DeleteSLO(context.Context, string) error { return nil }
func (f *fakeSLOService) EvaluateSLO(context.Context, string) (*models.SLOStatus, error) { return f.resp, nil }
func (f *fakeSLOService) GetSLOStatus(context.Context, string) (*models.SLOStatus, error) { return f.resp, nil }
func (f *fakeSLOService) GetSLOHistory(context.Context, string, time.Time, time.Time) ([]*models.SLOHistory, error) { return nil, nil }

type fakeQueries struct{}

func (f *fakeQueries) SearchTraces(ctx context.Context, service, operation string, minDur, maxDur time.Duration, errorOnly bool, userQuery string, start, end time.Time, limit int) ([]traces.TraceSummary, error) {
	return []traces.TraceSummary{{TraceID: "t1", Service: "svc", Operation: "op", DurationMs: 10, Error: false}}, nil
}
func (f *fakeQueries) GetTrace(ctx context.Context, id string) (*traces.TraceDetail, error) {
	return &traces.TraceDetail{TraceID: id, Spans: []json.RawMessage{[]byte(`{"span":"mock"}`)}}, nil
}
func (f *fakeQueries) SearchLogs(ctx context.Context, q string, start, end time.Time, limit int, direction string) ([]logs.LogEntry, error) {
	return []logs.LogEntry{{Timestamp: time.Now(), Line: "l", Labels: map[string]string{"k":"v"}}}, nil
}
func (f *fakeQueries) Golden(ctx context.Context, service string, window time.Duration) (*services.GoldenSignals, error) {
	return &services.GoldenSignals{RequestRate: 1.2, ErrorRate: 0.01, P95Latency: 0.2}, nil
}

func dialBufConn(t *testing.T, lis *bufconn.Listener) *grpc.ClientConn {
	t.Helper()
	RegisterJSONCodec()
	srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier{scopes: []string{"observability:read"}})))
	RegisterObservabilityServer(srv, NewObservabilityServer(&fakeSLOService{resp: &models.SLOStatus{SLOID: "s1", Compliance: 99.0, BurnRate: 0.1, RemainingBudget: 1.0}}, nil))
	go func(){ _ = srv.Serve(lis) }()
	dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
	cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer), grpc.WithDefaultCallOptions(grpc.ForceCodec(jsonCodec{})))
	if err != nil { t.Fatalf("dial: %v", err) }
	return cc
}

func TestGRPC_GetSLOStatus_OK(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	defer lis.Close()
	cc := dialBufConn(t, lis)
	defer cc.Close()

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer token"))
	var out GetSLOStatusResponse
	if err := cc.Invoke(ctx, "/"+serviceName+"/GetSLOStatus", &GetSLOStatusRequest{SloId: "s1"}, &out); err != nil {
		t.Fatalf("invoke: %v", err)
	}
	if out.SloId != "s1" || out.Compliance == 0 { t.Fatalf("bad response: %+v", out) }
}

func TestGRPC_SearchTraces_OK(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	defer lis.Close()
	RegisterJSONCodec()
	srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier{scopes: []string{"observability:read"}})))
	RegisterObservabilityServer(srv, NewObservabilityServer(&fakeSLOService{}, &fakeQueries{}))
	go func(){ _ = srv.Serve(lis) }()
	dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
	cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer), grpc.WithDefaultCallOptions(grpc.ForceCodec(jsonCodec{})))
	if err != nil { t.Fatalf("dial: %v", err) }
	defer cc.Close()
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer token"))
	var out SearchTracesResponse
	if err := cc.Invoke(ctx, "/"+serviceName+"/SearchTraces", &SearchTracesRequest{Service: "svc", Limit: 1}, &out); err != nil {
		t.Fatalf("invoke: %v", err)
	}
	if len(out.Traces) != 1 || out.Traces[0].TraceID != "t1" {
		t.Fatalf("unexpected traces: %+v", out.Traces)
	}
}

func TestGRPC_GetTrace_And_SearchLogs_And_Golden_OK(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	defer lis.Close()
	RegisterJSONCodec()
	srv := grpc.NewServer(grpc.UnaryInterceptor(UnaryAuthInterceptor(&fakeVerifier{scopes: []string{"observability:read"}})))
	RegisterObservabilityServer(srv, NewObservabilityServer(&fakeSLOService{}, &fakeQueries{}))
	go func(){ _ = srv.Serve(lis) }()
	dialer := func(ctx context.Context, _ string) (net.Conn, error) { return lis.Dial() }
	cc, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer), grpc.WithDefaultCallOptions(grpc.ForceCodec(jsonCodec{})))
	if err != nil { t.Fatalf("dial: %v", err) }
	defer cc.Close()
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", "Bearer token"))
	// GetTrace
	var gt GetTraceResponse
	if err := cc.Invoke(ctx, "/"+serviceName+"/GetTrace", &GetTraceRequest{TraceId:"abc"}, &gt); err != nil { t.Fatalf("gettrace: %v", err) }
	if gt.TraceId != "abc" || len(gt.Spans)==0 { t.Fatalf("bad gettrace resp: %+v", gt) }
	// SearchLogs
	var sl SearchLogsResponse
	if err := cc.Invoke(ctx, "/"+serviceName+"/SearchLogs", &SearchLogsRequest{Q:"{service=\"svc\"}"}, &sl); err != nil { t.Fatalf("searchlogs: %v", err) }
	if len(sl.Entries)==0 || sl.Entries[0].Line=="" { t.Fatalf("bad logs: %+v", sl) }
	// Golden
	var g GoldenResponse
	if err := cc.Invoke(ctx, "/"+serviceName+"/Golden", &GoldenRequest{Service:"svc", Window:"5m"}, &g); err != nil { t.Fatalf("golden: %v", err) }
	if g.RequestRate==0 || g.P95Latency==0 { t.Fatalf("bad golden: %+v", g) }
}
