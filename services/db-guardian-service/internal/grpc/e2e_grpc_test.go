package grpcserver

import (
    "context"
    "net"
    "sync"
    "testing"
    "time"

    dbguardianv1 "example.com/lma/db-guardian-service/dbguardian/v1"
    "example.com/lma/db-guardian-service/internal/analyzer"
    "example.com/lma/db-guardian-service/internal/auth"
    "example.com/lma/db-guardian-service/internal/dto"
    "example.com/lma/db-guardian-service/internal/models"
    "example.com/lma/db-guardian-service/internal/server"
    "example.com/lma/db-guardian-service/internal/service"
    "github.com/prometheus/client_golang/prometheus/testutil"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
)

// --- fakes

type e2eFakeAuth struct{ scopes []string }
func (f *e2eFakeAuth) Verify(_ context.Context, bearer string, required []string) (*auth.Claims, error) {
    if bearer == "Bearer ok" && e2eHasAllScopes(f.scopes, required) { return &auth.Claims{Subject: "u"}, nil }
    return nil, status.Error(codes.PermissionDenied, "forbidden")
}

type e2eTokenLimiter struct{ mu sync.Mutex; budget map[string]int }
func newTokenLimiter(initial int) *e2eTokenLimiter { return &e2eTokenLimiter{budget: map[string]int{"u": initial}} }
func (l *e2eTokenLimiter) Allow(_ context.Context, key string, cost int) (bool, error) {
    l.mu.Lock(); defer l.mu.Unlock()
    b, ok := l.budget[key]; if !ok { b = 0 }
    if b < cost { return false, nil }
    l.budget[key] = b - cost
    return true, nil
}

type e2eFakePolicy struct{}
func (f *e2eFakePolicy) GetPolicy(ctx context.Context, projectID string) (*models.RolePolicy, error) {
    return &models.RolePolicy{ProjectID: projectID, SpecYAML: "spec: v1", Version: 1}, nil
}
func (f *e2eFakePolicy) UpdatePolicy(ctx context.Context, projectID, yaml string) (*models.RolePolicy, error) {
    return &models.RolePolicy{ProjectID: projectID, SpecYAML: yaml, Version: 2}, nil
}

type e2eFakeRecs struct{}
func (f *e2eFakeRecs) Get(ctx context.Context, projectID string, onlyUnapplied bool) ([]analyzer.IndexRecommendation, *analyzer.RoleAnalysisResult, error) {
    idx := []analyzer.IndexRecommendation{{TableName: "t", Columns: []string{"c"}, BenefitScore: 5, Reason: "x"}}
    role := &analyzer.RoleAnalysisResult{RolesAnalyzed: 1}
    return idx, role, nil
}

type e2eFakeDrift struct{}
func (f *e2eFakeDrift) Check(ctx context.Context, projectID string) (*dto.DriftResponse, error) {
    return &dto.DriftResponse{MissingIndexes: []string{"t:c"}}, nil
}

// start a real grpc server listening on an ephemeral port
func startTestGRPC(t *testing.T, a server.Authenticator, rl server.RateLimiter) (addr string, stop func()) {
    t.Helper()
    deps := &server.Dependencies{
        Logger:      nil,
        PolicyMgr:   &e2eFakePolicy{},
        RecsProvider:&e2eFakeRecs{},
        DriftCheck:  &e2eFakeDrift{},
        Authenticator: a,
        RateLimiter: rl,
    }
    lis, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil { t.Fatalf("listen: %v", err) }
    s := NewGRPCServer(deps)
    ctx, cancel := context.WithCancel(context.Background())
    go func(){ _ = StartGRPC(ctx, lis, s) }()
    return lis.Addr().String(), func(){ cancel(); s.GracefulStop(); _ = lis.Close() }
}

func clientWithAuth(t *testing.T, addr string) (dbguardianv1.DbGuardianServiceClient, func(context.Context) context.Context, func()) {
    t.Helper()
    cc, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { t.Fatalf("dial: %v", err) }
    c := dbguardianv1.NewDbGuardianServiceClient(cc)
    withAuth := func(ctx context.Context) context.Context { return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer ok") }
    return c, withAuth, func(){ _ = cc.Close() }
}

// --- tests

func TestE2E_GetPolicy_RequiresReadScope(t *testing.T) {
    addr, stop := startTestGRPC(t, &e2eFakeAuth{scopes: []string{}}, nil)
    defer stop()
    c, withAuth, closeFn := clientWithAuth(t, addr); defer closeFn()
    // Missing read → denied
    _, err := c.GetPolicy(withAuth(context.Background()), &dbguardianv1.GetPolicyRequest{ProjectId: "p1"})
    if status.Code(err) != codes.PermissionDenied { t.Fatalf("expected PermissionDenied, got %v", status.Code(err)) }

    // With read → OK
    stop()
    addr2, stop2 := startTestGRPC(t, &e2eFakeAuth{scopes: []string{"db.read"}}, nil)
    defer stop2()
    c2, withAuth2, close2 := clientWithAuth(t, addr2); defer close2()
    resp, err := c2.GetPolicy(withAuth2(context.Background()), &dbguardianv1.GetPolicyRequest{ProjectId: "p1"})
    if err != nil { t.Fatalf("unexpected: %v", err) }
    if resp.GetProjectId() != "p1" { t.Fatalf("bad resp: %+v", resp) }
}

func TestE2E_UpdatePolicy_RequiresWriteScope(t *testing.T) {
    addr, stop := startTestGRPC(t, &e2eFakeAuth{scopes: []string{"db.read"}}, nil)
    defer stop()
    c, withAuth, closeFn := clientWithAuth(t, addr); defer closeFn()
    _, err := c.UpdatePolicy(withAuth(context.Background()), &dbguardianv1.UpdatePolicyRequest{ProjectId: "p1", SpecYaml: "x: 1"})
    if status.Code(err) != codes.PermissionDenied { t.Fatalf("expected PermissionDenied, got %v", status.Code(err)) }

    stop()
    addr2, stop2 := startTestGRPC(t, &e2eFakeAuth{scopes: []string{"db.write"}}, nil)
    defer stop2()
    c2, withAuth2, close2 := clientWithAuth(t, addr2); defer close2()
    r, err := c2.UpdatePolicy(withAuth2(context.Background()), &dbguardianv1.UpdatePolicyRequest{ProjectId: "p1", SpecYaml: "x: 2"})
    if err != nil { t.Fatalf("unexpected: %v", err) }
    if r.GetVersion() != 2 { t.Fatalf("expected version 2, got %d", r.GetVersion()) }
}

func TestE2E_RateLimit_DeniesOverBudget(t *testing.T) {
    rl := newTokenLimiter(1) // allow one read (cost 1)
    addr, stop := startTestGRPC(t, &e2eFakeAuth{scopes: []string{"db.read"}}, rl)
    defer stop()
    c, withAuth, closeFn := clientWithAuth(t, addr); defer closeFn()
    ctx := withAuth(context.Background())
    _, err := c.GetPolicy(ctx, &dbguardianv1.GetPolicyRequest{ProjectId: "p1"})
    if err != nil { t.Fatalf("unexpected: %v", err) }
    _, err = c.GetPolicy(ctx, &dbguardianv1.GetPolicyRequest{ProjectId: "p1"})
    if status.Code(err) != codes.ResourceExhausted { t.Fatalf("expected ResourceExhausted, got %v", status.Code(err)) }
}

func TestE2E_Metrics_EmittedForRPC(t *testing.T) {
    addr, stop := startTestGRPC(t, &e2eFakeAuth{scopes: []string{"db.read"}}, nil)
    defer stop()
    c, withAuth, closeFn := clientWithAuth(t, addr); defer closeFn()
    method := "/dbguardian.v1.DbGuardianService/GetPolicy"
    before := testutil.ToFloat64(grpcRequestsTotal.WithLabelValues(method, codes.OK.String()))
    _, _ = c.GetPolicy(withAuth(context.Background()), &dbguardianv1.GetPolicyRequest{ProjectId: "p1"})
    // slight wait to allow metrics to flush
    time.Sleep(5 * time.Millisecond)
    after := testutil.ToFloat64(grpcRequestsTotal.WithLabelValues(method, codes.OK.String()))
    if after < before+1 { t.Fatalf("expected counter increase, before=%v after=%v", before, after) }
}

// Ensure no import pruning of service package accidentally
var _ = service.NewPolicyService

func e2eHasAllScopes(have, need []string) bool {
    if len(need) == 0 { return true }
    set := make(map[string]struct{}, len(have))
    for _, s := range have { set[s] = struct{}{} }
    for _, n := range need { if _, ok := set[n]; !ok { return false } }
    return true
}
