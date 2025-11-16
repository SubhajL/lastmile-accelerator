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
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
    "google.golang.org/grpc/test/bufconn"
)

type claimsVerifier struct{ claims handlers.Claims }

func (v *claimsVerifier) Verify(ctx context.Context, token string) (*handlers.Claims, error) { return &v.claims, nil }

func startRBACBufGRPC(t *testing.T, claims handlers.Claims) (*grpc.ClientConn, func()) {
    t.Helper()
    log := appLogger.New("test", "info", nil)
    lis := bufconn.Listen(1 << 20)
    secretsRepo := repository.NewSecretsRepository(nil)
    parityRepo := repository.NewParityRepository()
    secretsSvc := service.NewSecretsService(nil, secretsRepo, nil, nil)
    paritySvc := service.NewParityService(parityRepo, secretsRepo, nil)
    ver := &claimsVerifier{claims: claims}
    interceptors := []grpc.UnaryServerInterceptor{
        unaryAuth(ver),
        unaryRBACAugment(),
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

func TestGRPC_RBAC_Admin_AllowsCompute(t *testing.T) {
    conn, stop := startRBACBufGRPC(t, handlers.Claims{TenantID:"t1", ProjectID:"p1", Roles:[]string{"admin"}})
    defer stop()
    md := metadata.Pairs("authorization", "Bearer ok")
    ctx := metadata.NewOutgoingContext(context.Background(), md)
    client := pb.NewSecretsEnvServiceClient(conn)
    if _, err := client.CheckEnvParity(ctx, &pb.CheckEnvParityRequest{ProjectId:"p1", BaseEnv:"dev", CompareEnv:"prod"}, grpc.ForceCodec(jsonCodec{})); err != nil {
        t.Fatalf("unexpected err: %v", err)
    }
}

func TestGRPC_RBAC_Auditor_DeniesCompute_AllowsRead(t *testing.T) {
    conn, stop := startRBACBufGRPC(t, handlers.Claims{TenantID:"t1", ProjectID:"p1", Roles:[]string{"auditor"}})
    defer stop()
    md := metadata.Pairs("authorization", "Bearer ok")
    ctx := metadata.NewOutgoingContext(context.Background(), md)
    client := pb.NewSecretsEnvServiceClient(conn)
    if _, err := client.ListSecrets(ctx, &pb.ListSecretsRequest{ProjectId:"p1", Environment:"dev", PageSize:1}, grpc.ForceCodec(jsonCodec{})); err != nil {
        t.Fatalf("list should pass: %v", err)
    }
    _, err := client.CheckEnvParity(ctx, &pb.CheckEnvParityRequest{ProjectId:"p1", BaseEnv:"dev", CompareEnv:"prod"}, grpc.ForceCodec(jsonCodec{}))
    if st, _ := status.FromError(err); st == nil || st.Code().String() != "PermissionDenied" { t.Fatalf("expected PermissionDenied, got %v", err) }
}
