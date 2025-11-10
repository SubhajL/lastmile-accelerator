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
	"example.com/lma/secrets-env-service/internal/security"
	"example.com/lma/secrets-env-service/internal/service"
	"example.com/lma/secrets-env-service/internal/vault"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

type fakeVerifier struct{ allow bool; claims handlers.Claims; err error }

func (f *fakeVerifier) Verify(ctx context.Context, token string) (*handlers.Claims, error) {
	if !f.allow { return nil, context.Canceled }
	return &f.claims, nil
}

func TestGRPC_RateLimit_ResourceExhausted(t *testing.T) {
	v := &vault.Client{}; v.SetTestMode(true)
	repo := repository.NewSecretsRepository(nil)
	secretsSvc := service.NewSecretsService(v, repo, nil, nil)
	paritySvc := service.NewParityService(repository.NewParityRepository(), repo, nil)
	ver := &fakeVerifier{allow: true, claims: handlers.Claims{TenantID: "t1", ProjectID: "p", Scopes: []string{"secrets:read"}}}
	lim := security.NewRateLimiter(1, 1)
	conn, stop := startBufGRPC(t, secretsSvc, paritySvc, ver, lim)
	defer stop()
	md := metadata.Pairs("authorization", "Bearer ok")
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	client := pb.NewSecretsEnvServiceClient(conn)
	// First should pass
	_, _ = client.ListSecrets(ctx, &pb.ListSecretsRequest{ProjectId: "p", Environment: "dev", PageSize: 1}, grpc.ForceCodec(jsonCodec{}))
	// Second immediately should hit limiter
	_, err := client.ListSecrets(ctx, &pb.ListSecretsRequest{ProjectId: "p", Environment: "dev", PageSize: 1}, grpc.ForceCodec(jsonCodec{}))
	if err == nil { t.Fatalf("expected rate limit error") }
}

func startBufGRPC(t *testing.T, secrets *service.SecretsService, parity *service.ParityService, ver handlers.TokenVerifier, limiter *security.RateLimiter) (*grpc.ClientConn, func()) {
	t.Helper()
	log := appLogger.New("test", "info", nil)
	lis := bufconn.Listen(1 << 20)
	interceptors := []grpc.UnaryServerInterceptor{
		unaryPanicRecovery(log),
		unaryRequestLogger(log),
		unaryAuth(ver),
	}
	if limiter != nil {
		interceptors = append(interceptors, unaryRateLimit(limiter))
	}
	interceptors = append(interceptors, unaryRequireScopes(map[string][]string{
		"/lma.secretsenv.v1.SecretsEnvService/GetSecret":   {"secrets:read"},
		"/lma.secretsenv.v1.SecretsEnvService/ListSecrets":  {"secrets:read"},
		"/lma.secretsenv.v1.SecretsEnvService/CheckEnvParity": {"parity:compute"},
	}))
	gs := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	svc := &server{secrets: secrets, parity: parity, log: log}
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

func TestGRPC_GetSecret_Success_ReturnsMetaAndValue(t *testing.T) {
	// Seed secret
	v := &vault.Client{}; v.SetTestMode(true)
	repo := repository.NewSecretsRepository(nil)
secretsSvc := service.NewSecretsService(v, repo, nil, nil)
	sec := &domain.Secret{ID: "1", TenantID: "t1", ProjectID: "p", Key: "K", Environment: "dev"}
	_ = repo.Create(context.Background(), sec)
	_ = v.WriteSecret(context.Background(), sec.VaultPath(), map[string]any{"x":"y"})

	paritySvc := service.NewParityService(repository.NewParityRepository(), repo, nil)
	ver := &fakeVerifier{allow: true, claims: handlers.Claims{TenantID: "t1", ProjectID: "p", Scopes: []string{"secrets:read"}}}
conn, stop := startBufGRPC(t, secretsSvc, paritySvc, ver, nil)
	defer stop()

	md := metadata.Pairs("authorization", "Bearer ok")
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	client := pb.NewSecretsEnvServiceClient(conn)
	req := &pb.GetSecretRequest{TenantId: "t1", ProjectId: "p", Key: "K", Environment: "dev"}
	resp, err := client.GetSecret(ctx, req, grpc.ForceCodec(jsonCodec{}))
	if err != nil { t.Fatalf("invoke err: %v", err) }
	if resp.Meta.Key != "K" || resp.Value["x"] != "y" { t.Fatalf("unexpected resp: %#v", resp) }
}

func TestGRPC_Auth_ScopeEnforced_GetSecretRequiresSecretsRead(t *testing.T) {
	v := &vault.Client{}; v.SetTestMode(true)
	repo := repository.NewSecretsRepository(nil)
secretsSvc := service.NewSecretsService(v, repo, nil, nil)
	paritySvc := service.NewParityService(repository.NewParityRepository(), repo, nil)

	verNo := &fakeVerifier{allow: true, claims: handlers.Claims{TenantID: "t1", ProjectID: "p", Scopes: []string{"secrets:write"}}}
conn, stop := startBufGRPC(t, secretsSvc, paritySvc, verNo, nil)
	defer stop()
	md := metadata.Pairs("authorization", "Bearer ok")
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	client := pb.NewSecretsEnvServiceClient(conn)
	req := &pb.GetSecretRequest{TenantId: "t1", ProjectId: "p", Key: "K", Environment: "dev"}
	_, err := client.GetSecret(ctx, req, grpc.ForceCodec(jsonCodec{}))
	if err == nil { t.Fatalf("expected permission denied") }
}

func TestGRPC_CheckEnvParity_Success_ReturnsMissingAndExtra(t *testing.T) {
	// dev:A,B; prod:B,C -> missing:A extra:C
	repo := repository.NewSecretsRepository(nil)
	_ = repo.Create(context.Background(), &domain.Secret{ID:"1", TenantID:"t1", ProjectID:"p", Key:"A", Environment:"dev"})
	_ = repo.Create(context.Background(), &domain.Secret{ID:"2", TenantID:"t1", ProjectID:"p", Key:"B", Environment:"dev"})
	_ = repo.Create(context.Background(), &domain.Secret{ID:"3", TenantID:"t1", ProjectID:"p", Key:"B", Environment:"prod"})
	_ = repo.Create(context.Background(), &domain.Secret{ID:"4", TenantID:"t1", ProjectID:"p", Key:"C", Environment:"prod"})

	paritySvc := service.NewParityService(repository.NewParityRepository(), repo, nil)
secretsSvc := service.NewSecretsService(&vault.Client{}, repo, nil, nil)
	ver := &fakeVerifier{allow: true, claims: handlers.Claims{TenantID: "t1", ProjectID: "p", Scopes: []string{"parity:compute"}}}
conn, stop := startBufGRPC(t, secretsSvc, paritySvc, ver, nil)
	defer stop()
	md := metadata.Pairs("authorization", "Bearer ok")
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	client := pb.NewSecretsEnvServiceClient(conn)
	req := &pb.CheckEnvParityRequest{ProjectId: "p", BaseEnv: "dev", CompareEnv: "prod"}
	resp, err := client.CheckEnvParity(ctx, req, grpc.ForceCodec(jsonCodec{}))
	if err != nil { t.Fatalf("invoke err: %v", err) }
	if len(resp.MissingKeys) != 1 || resp.MissingKeys[0] != "A" { t.Fatalf("missing: %#v", resp.MissingKeys) }
	if len(resp.ExtraKeys) != 1 || resp.ExtraKeys[0] != "C" { t.Fatalf("extra: %#v", resp.ExtraKeys) }
}