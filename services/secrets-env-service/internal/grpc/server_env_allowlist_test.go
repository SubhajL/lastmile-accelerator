package grpcapi

import (
    "context"
    "net"
    "testing"

    pb "example.com/lma/secrets-env-service/internal/grpc/gen/secretsenvv1"
    "example.com/lma/secrets-env-service/internal/handlers"
    appLogger "example.com/lma/secrets-env-service/internal/logger"
    "example.com/lma/secrets-env-service/internal/repository"
    "example.com/lma/secrets-env-service/internal/service"
    "github.com/prometheus/client_golang/prometheus"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/test/bufconn"
)

type allowlistVerifier struct{ claims handlers.Claims }

func (v *allowlistVerifier) Verify(ctx context.Context, token string) (*handlers.Claims, error) { return &v.claims, nil }

func startBufGRPCWithAllowlist(t *testing.T, allowed []string) (*grpc.ClientConn, func()) {
    t.Helper()
    log := appLogger.New("test", "info", nil)
    lis := bufconn.Listen(1 << 20)
    secretsRepo := repository.NewSecretsRepository(nil)
    parityRepo := repository.NewParityRepository()
    secretsSvc := service.NewSecretsService(nil, secretsRepo, nil, nil)
    paritySvc := service.NewParityService(parityRepo, secretsRepo, nil)
    ver := &allowlistVerifier{claims: handlers.Claims{TenantID: "t1", ProjectID: "p1", Scopes: []string{"secrets:read","parity:compute"}}}
    interceptors := []grpc.UnaryServerInterceptor{
        unaryEnvAllowlist(allowed),
        unaryAuth(ver),
        unaryRequireScopes(map[string][]string{
            "/lma.secretsenv.v1.SecretsEnvService/GetSecret":   {"secrets:read"},
            "/lma.secretsenv.v1.SecretsEnvService/ListSecrets":  {"secrets:read"},
            "/lma.secretsenv.v1.SecretsEnvService/CheckEnvParity": {"parity:compute"},
        }),
    }
    _ = prometheus.NewRegistry() // silence import if unused
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

func TestGRPC_EnvAllowlist_InvalidEnv(t *testing.T) {
    conn, stop := startBufGRPCWithAllowlist(t, []string{"dev"})
    defer stop()
    md := metadata.Pairs("authorization", "Bearer ok")
    ctx := metadata.NewOutgoingContext(context.Background(), md)
    client := pb.NewSecretsEnvServiceClient(conn)
    // ListSecrets with disallowed env
    _, err := client.ListSecrets(ctx, &pb.ListSecretsRequest{ProjectId:"p1", Environment:"stage", PageSize:1}, grpc.ForceCodec(jsonCodec{}))
    if st, _ := status.FromError(err); st == nil || st.Code().String() != "InvalidArgument" { t.Fatalf("expected InvalidArgument, got %v", err) }
}
