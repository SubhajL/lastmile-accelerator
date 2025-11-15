package grpcapi

import (
    "context"
    "net"
    "testing"

    "example.com/lma/secrets-env-service/internal/domain"
    pb "example.com/lma/secrets-env-service/internal/grpc/gen/secretsenvv1"
    "example.com/lma/secrets-env-service/internal/handlers"
    appLogger "example.com/lma/secrets-env-service/internal/logger"
    "example.com/lma/secrets-env-service/internal/repository"
    "example.com/lma/secrets-env-service/internal/service"
    "github.com/prometheus/client_golang/prometheus"
    dto "github.com/prometheus/client_model/go"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/test/bufconn"
)

type metricsFakeVerifier struct{ claims handlers.Claims }

func (f *metricsFakeVerifier) Verify(ctx context.Context, token string) (*handlers.Claims, error) {
    return &f.claims, nil
}

func startBufGRPCWithMetrics(t *testing.T, reg *prometheus.Registry, claims handlers.Claims) (*grpc.ClientConn, func()) {
    t.Helper()
    log := appLogger.New("test", "info", nil)
    lis := bufconn.Listen(1 << 20)
    // Minimal services
    repo := repository.NewSecretsRepository(nil)
    secretsSvc := service.NewSecretsService(nil, repo, nil, nil)
    paritySvc := service.NewParityService(repository.NewParityRepository(), repo, nil)
    // Register a secret to make GetSecret succeed when allowed
    _ = repo.Create(context.Background(), &domain.Secret{ID: "1", TenantID: claims.TenantID, ProjectID: claims.ProjectID, Key: "K", Environment: "dev"})

    ver := &metricsFakeVerifier{claims: claims}
    interceptors := []grpc.UnaryServerInterceptor{
        unaryGRPCMetrics(reg),
        unaryAuth(ver),
        unaryRequireScopes(map[string][]string{
            "/lma.secretsenv.v1.SecretsEnvService/GetSecret":   {"secrets:read"},
            "/lma.secretsenv.v1.SecretsEnvService/ListSecrets":  {"secrets:read"},
            "/lma.secretsenv.v1.SecretsEnvService/CheckEnvParity": {"parity:compute"},
        }),
    }
    gs := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
    svc := &server{secrets: secretsSvc, parity: paritySvc, log: log}
    pb.RegisterSecretsEnvServiceServer(gs, svc)
    go func() { _ = gs.Serve(lis) }()
    dialer := func(ctx context.Context, s string) (net.Conn, error) { return lis.Dial() }
    conn, err := grpc.DialContext(context.Background(), "bufnet",
        grpc.WithContextDialer(dialer),
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithDefaultCallOptions(grpc.ForceCodec(jsonCodec{})),
    )
    if err != nil { t.Fatalf("dial err: %v", err) }
    stop := func() { gs.GracefulStop(); _ = lis.Close(); _ = conn.Close() }
    return conn, stop
}

func gatherFamily(reg *prometheus.Registry, name string) *dto.MetricFamily {
    fams, err := reg.Gather()
    if err != nil { return nil }
    for _, f := range fams { if f.GetName() == name { return f } }
    return nil
}

func TestGRPCMetrics_Counter_Increments_OnOK(t *testing.T) {
    reg := prometheus.NewRegistry()
    claims := handlers.Claims{TenantID: "t1", ProjectID: "p1", Scopes: []string{"secrets:read"}}
    conn, stop := startBufGRPCWithMetrics(t, reg, claims)
    defer stop()
    md := metadata.Pairs("authorization", "Bearer ok")
    ctx := metadata.NewOutgoingContext(context.Background(), md)
    client := pb.NewSecretsEnvServiceClient(conn)
    _, _ = client.ListSecrets(ctx, &pb.ListSecretsRequest{ProjectId: "p1", Environment: "dev", PageSize: 1})
    fam := gatherFamily(reg, "grpc_requests_total")
    if fam == nil { t.Fatalf("grpc_requests_total not found") }
    found := false
    for _, m := range fam.GetMetric() {
        var method, code string
        for _, lp := range m.GetLabel() {
            switch lp.GetName() { case "method": method = lp.GetValue(); case "code": code = lp.GetValue() }
        }
        if method == "/lma.secretsenv.v1.SecretsEnvService/ListSecrets" && code == "OK" && m.GetCounter().GetValue() >= 1 {
            found = true
            break
        }
    }
    if !found { t.Fatalf("expected counter for ListSecrets OK >=1") }
}

func TestGRPCMetrics_Histogram_Records_OnOK(t *testing.T) {
    reg := prometheus.NewRegistry()
    claims := handlers.Claims{TenantID: "t1", ProjectID: "p1", Scopes: []string{"secrets:read"}}
    conn, stop := startBufGRPCWithMetrics(t, reg, claims)
    defer stop()
    md := metadata.Pairs("authorization", "Bearer ok")
    ctx := metadata.NewOutgoingContext(context.Background(), md)
    client := pb.NewSecretsEnvServiceClient(conn)
    _, _ = client.ListSecrets(ctx, &pb.ListSecretsRequest{ProjectId: "p1", Environment: "dev", PageSize: 1})
    fam := gatherFamily(reg, "grpc_request_duration_seconds")
    if fam == nil { t.Fatalf("grpc_request_duration_seconds not found") }
    samples := uint64(0)
    for _, m := range fam.GetMetric() { samples += m.GetHistogram().GetSampleCount() }
    if samples == 0 { t.Fatalf("expected histogram samples > 0") }
}

func TestGRPCMetrics_Counter_Labels_OnPermissionDenied(t *testing.T) {
    reg := prometheus.NewRegistry()
    // Missing secrets:read scope -> permission denied
    claims := handlers.Claims{TenantID: "t1", ProjectID: "p1", Scopes: []string{"secrets:write"}}
    conn, stop := startBufGRPCWithMetrics(t, reg, claims)
    defer stop()
    md := metadata.Pairs("authorization", "Bearer ok")
    ctx := metadata.NewOutgoingContext(context.Background(), md)
    client := pb.NewSecretsEnvServiceClient(conn)
    _, _ = client.GetSecret(ctx, &pb.GetSecretRequest{TenantId: "t1", ProjectId: "p1", Key: "K", Environment: "dev"})
    fam := gatherFamily(reg, "grpc_requests_total")
    if fam == nil { t.Fatalf("grpc_requests_total not found") }
    found := false
    for _, m := range fam.GetMetric() {
        var method, code string
        for _, lp := range m.GetLabel() {
            switch lp.GetName() { case "method": method = lp.GetValue(); case "code": code = lp.GetValue() }
        }
        if method == "/lma.secretsenv.v1.SecretsEnvService/GetSecret" && code == "PermissionDenied" && m.GetCounter().GetValue() >= 1 {
            found = true
            break
        }
    }
    if !found { t.Fatalf("expected counter for GetSecret PermissionDenied >=1") }
}

func TestGRPCMetrics_RegisterTwice_AlreadyRegisteredReused(t *testing.T) {
    reg := prometheus.NewRegistry()
    claims := handlers.Claims{TenantID: "t1", ProjectID: "p1", Scopes: []string{"secrets:read"}}
    c1, stop1 := startBufGRPCWithMetrics(t, reg, claims)
    defer stop1()
    c2, stop2 := startBufGRPCWithMetrics(t, reg, claims)
    defer stop2()
    _ = c2
    // Make a call to ensure metrics have samples and collectors are in registry
    md := metadata.Pairs("authorization", "Bearer ok")
    ctx := metadata.NewOutgoingContext(context.Background(), md)
    client := pb.NewSecretsEnvServiceClient(c1)
    _, _ = client.ListSecrets(ctx, &pb.ListSecretsRequest{ProjectId: "p1", Environment: "dev", PageSize: 1})
    // If we reach here without panic, and registry has the counter family, we pass
    if gatherFamily(reg, "grpc_requests_total") == nil { t.Fatalf("counter family not found after double register") }
}
