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
	"example.com/lma/secrets-env-service/internal/vault"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

func TestGrpcMetrics_RecordsCounterAndDuration(t *testing.T) {
	reg := prometheus.NewRegistry()

	v := &vault.Client{}
	v.SetTestMode(true)
	repo := repository.NewSecretsRepository(nil)
	secretsSvc := service.NewSecretsService(v, repo, nil, nil)
	paritySvc := service.NewParityService(repository.NewParityRepository(), repo, nil)
	ver := &fakeVerifier{allow: true, claims: handlers.Claims{TenantID: "t1", ProjectID: "p", Scopes: []string{"secrets:read"}}}

	log := appLogger.New("test", "info", nil)
	lis := bufconn.Listen(1 << 20)
	interceptors := []grpc.UnaryServerInterceptor{
		unaryPanicRecovery(log),
		unaryGRPCMetrics(reg),
		unaryAuth(ver),
		unaryRequireScopes(map[string][]string{
			"/lma.secretsenv.v1.SecretsEnvService/GetSecret":      {"secrets:read"},
			"/lma.secretsenv.v1.SecretsEnvService/ListSecrets":    {"secrets:read"},
			"/lma.secretsenv.v1.SecretsEnvService/CheckEnvParity": {"parity:compute"},
		}),
	}

	gs := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	pb.RegisterSecretsEnvServiceServer(gs, &server{secrets: secretsSvc, parity: paritySvc, log: log})

	go func() { _ = gs.Serve(lis) }()
	t.Cleanup(func() { gs.GracefulStop(); _ = lis.Close() })

	dialer := func(ctx context.Context, s string) (net.Conn, error) { return lis.Dial() }
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(jsonCodec{})),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	md := metadata.Pairs("authorization", "Bearer ok")
	ctx := metadata.NewOutgoingContext(context.Background(), md)
	client := pb.NewSecretsEnvServiceClient(conn)
	_, err = client.ListSecrets(ctx, &pb.ListSecretsRequest{ProjectId: "p", Environment: "dev", PageSize: 1}, grpc.ForceCodec(jsonCodec{}))
	require.NoError(t, err)

	families, err := reg.Gather()
	require.NoError(t, err)

	methodLabel := "/lma.secretsenv.v1.SecretsEnvService/ListSecrets"

	counterFound := false
	histFound := false
	for _, f := range families {
		switch f.GetName() {
		case "grpc_requests_total":
			for _, m := range f.GetMetric() {
				labels := map[string]string{}
				for _, lp := range m.GetLabel() {
					labels[lp.GetName()] = lp.GetValue()
				}

				if labels["method"] == methodLabel &&
					labels["status"] == "OK" &&
					m.GetCounter().GetValue() >= 1 {
					counterFound = true
					break
				}
			}
		case "grpc_request_duration_seconds":
			for _, m := range f.GetMetric() {
				labels := map[string]string{}
				for _, lp := range m.GetLabel() {
					labels[lp.GetName()] = lp.GetValue()
				}

				if labels["method"] == methodLabel && m.GetHistogram().GetSampleCount() > 0 {
					histFound = true
					break
				}
			}
		}
	}

	assert.True(t, counterFound, "expected grpc_requests_total for method=%s status=OK", methodLabel)
	assert.True(t, histFound, "expected grpc_request_duration_seconds for method=%s", methodLabel)
}
