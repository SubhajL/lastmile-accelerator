package grpcapi

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"strings"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	pb "example.com/lma/secrets-env-service/internal/grpc/gen/secretsenvv1"
	"example.com/lma/secrets-env-service/internal/handlers"
	"example.com/lma/secrets-env-service/internal/logger"
	"example.com/lma/secrets-env-service/internal/security"
	"example.com/lma/secrets-env-service/internal/service"
	"example.com/lma/secrets-env-service/internal/events"
    "github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// RPC request/response types (JSON-encoded via jsonCodec)

type GetSecretRequest struct {
	TenantID    string `json:"tenant_id"`
	ProjectID   string `json:"project_id"`
	Key         string `json:"key"`
	Environment string `json:"environment"`
}

type SecretMeta struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Environment string    `json:"environment"`
	CreatedAt   time.Time `json:"created_at"`
}

type GetSecretResponse struct {
	Meta  SecretMeta            `json:"meta"`
	Value map[string]any        `json:"value"`
}

type ListSecretsRequest struct {
	ProjectID   string `json:"project_id"`
	Environment string `json:"environment"`
	PageSize    int32  `json:"page_size"`
	PageToken   string `json:"page_token"`
}

type ListSecretsResponse struct {
	Items         []SecretMeta `json:"items"`
	NextPageToken string      `json:"next_page_token"`
}

type CheckEnvParityRequest struct {
	ProjectID string `json:"project_id"`
	BaseEnv   string `json:"base_env"`
	CompareEnv string `json:"compare_env"`
}

type CheckEnvParityResponse struct {
	ProjectID     string    `json:"project_id"`
	ScanTimestamp time.Time `json:"scan_timestamp"`
	MissingKeys   []string  `json:"missing_keys"`
	ExtraKeys     []string  `json:"extra_keys"`
	HasDrift      bool      `json:"has_drift"`
}

// Server wiring

type server struct {
	pb.UnimplementedSecretsEnvServiceServer
	secrets *service.SecretsService
	parity  *service.ParityService
	log     zerolog.Logger
}

// StartGRPCServer starts the gRPC server on addr and blocks until ctx is canceled.
func StartGRPCServer(ctx context.Context, addr string, secrets *service.SecretsService, parity *service.ParityService, verifier handlers.TokenVerifier, log zerolog.Logger, tlsConfig *tls.Config, allowed []string, enableHealth bool, enableReflection bool) error {
	// Interceptor chain: recovery -> logging -> auth -> scope enforcement
	svc := &server{secrets: secrets, parity: parity, log: log}

	requiredScopes := map[string][]string{
		"/lma.secretsenv.v1.SecretsEnvService/GetSecret":   {"secrets:read"},
		"/lma.secretsenv.v1.SecretsEnvService/ListSecrets":  {"secrets:read"},
		"/lma.secretsenv.v1.SecretsEnvService/CheckEnvParity": {"parity:compute"},
	}

chain := grpc.ChainUnaryInterceptor(
        unaryEnvAllowlist(allowed),
        unaryGRPCMetrics(prometheus.DefaultRegisterer),
		unaryPanicRecovery(log),
		unaryTraceContext(),
		unaryOtelTracing(),
		unaryRequestLogger(log),
		unaryAuth(verifier),
		unaryRBACAugment(),
		unaryRateLimit(security.NewRateLimiter(20, 40)),
		unaryRequireScopes(requiredScopes),
	)
var opts []grpc.ServerOption
opts = append(opts, chain)
if tlsConfig != nil {
	opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
}
gs := grpc.NewServer(opts...)
pb.RegisterSecretsEnvServiceServer(gs, svc)

    // Optional health and reflection
    var hs *health.Server
    if enableHealth {
        var cleanup func()
        hs, cleanup = registerHealth(gs, ctx)
        _ = cleanup
        // Mark overall server as SERVING
        hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
    }
    if enableReflection {
        registerReflection(gs)
    }

	lis, err := net.Listen("tcp", addr)
	if err != nil { return err }
	go func() {
		<-ctx.Done()
		gs.GracefulStop()
		_ = lis.Close()
	}()
	return gs.Serve(lis)
}

// registerHealth registers the health server and ties its shutdown to ctx.
func registerHealth(gs *grpc.Server, ctx context.Context) (*health.Server, func()) {
    hs := health.NewServer()
    healthpb.RegisterHealthServer(gs, hs)
    cleanup := func() { hs.Shutdown() }
    go func() { <-ctx.Done(); hs.Shutdown() }()
    return hs, cleanup
}

// registerReflection registers the gRPC reflection service.
func registerReflection(gs *grpc.Server) {
    reflection.Register(gs)
}

// Handlers (pb interfaces)

func (s *server) GetSecret(ctx context.Context, in *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	claims, ok := handlers.ClaimsFromContext(ctx)
	if !ok { return nil, status.Error(codes.Unauthenticated, "missing claims") }
	sec, val, err := s.secrets.GetSecret(ctx, claims.TenantID, in.ProjectId, in.Key, in.Environment)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") { return nil, status.Error(codes.NotFound, "not found") }
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetSecretResponse{Meta: toPBSecretMeta(sec), Value: val}, nil
}

func (s *server) ListSecrets(ctx context.Context, in *pb.ListSecretsRequest) (*pb.ListSecretsResponse, error) {
	limit := int(in.PageSize)
	if limit <= 0 { limit = 50 }
	items, next, err := s.secrets.ListSecrets(ctx, in.ProjectId, in.Environment, limit, in.PageToken)
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	out := make([]*pb.SecretMeta, 0, len(items))
	for _, it := range items { out = append(out, toPBSecretMeta(it)) }
	return &pb.ListSecretsResponse{Items: out, NextPageToken: next}, nil
}

func (s *server) CheckEnvParity(ctx context.Context, in *pb.CheckEnvParityRequest) (*pb.CheckEnvParityResponse, error) {
	res, err := s.parity.CheckParity(ctx, in.ProjectId, in.BaseEnv, in.CompareEnv)
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	return &pb.CheckEnvParityResponse{
		ProjectId:     res.ProjectID,
		MissingKeys:   res.MissingKeys,
		ExtraKeys:     res.ExtraKeys,
		HasDrift:      res.HasDrift(),
	}, nil
}


func toPBSecretMeta(s *domain.Secret) *pb.SecretMeta {
	return &pb.SecretMeta{Id: s.ID, Key: s.Key, Environment: s.Environment}
}

// Interceptors

func unaryTraceContext() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		tp := ""
		if md != nil {
			vals := md.Get("traceparent")
			if len(vals) > 0 { tp = vals[0] }
		}
		ctx = events.WithTraceparent(ctx, tp)
		return handler(ctx, req)
	}
}

func unaryOtelTracing() grpc.UnaryServerInterceptor {
	tr := otel.Tracer("grpc-server")
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, span := tr.Start(ctx, info.FullMethod)
		defer span.End()
		resp, err := handler(ctx, req)
		if err != nil {
			span.SetAttributes(attribute.String("grpc.code", status.Code(err).String()))
		} else {
			span.SetAttributes(attribute.String("grpc.code", codes.OK.String()))
		}
		return resp, err
	}
}

func unaryPanicRecovery(log zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.SafeLogError(log, errors.New("panic"), "grpc panic recovered")
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

func unaryRequestLogger(log zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		dur := time.Since(start)
		code := codes.OK
		if st, ok := status.FromError(err); ok { code = st.Code() }
		log.Info().Str("method", info.FullMethod).Str("grpc_code", code.String()).Dur("duration_ms", dur).Msg("grpc")
		return resp, err
	}
}

func unaryAuth(verifier handlers.TokenVerifier) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		vals := md.Get("authorization")
		if len(vals) == 0 || !strings.HasPrefix(strings.ToLower(vals[0]), "bearer ") {
			return nil, status.Error(codes.Unauthenticated, "missing bearer token")
		}
		tok := strings.TrimPrefix(vals[0], "Bearer ")
		tok = strings.TrimPrefix(tok, "bearer ")
		claims, err := verifier.Verify(ctx, tok)
		if err != nil || claims == nil { return nil, status.Error(codes.Unauthenticated, "invalid token") }
		ctx = handlers.WithClaims(ctx, *claims)
		return handler(ctx, req)
	}
}

func unaryRequireScopes(required map[string][]string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		reqScopes := required[info.FullMethod]
		if len(reqScopes) == 0 { return handler(ctx, req) }
		claims, ok := handlers.ClaimsFromContext(ctx)
		if !ok { return nil, status.Error(codes.PermissionDenied, "missing claims") }
		set := map[string]struct{}{}
		for _, s := range claims.Scopes { set[s] = struct{}{} }
		for _, need := range reqScopes { if _, ok := set[need]; !ok { return nil, status.Error(codes.PermissionDenied, "insufficient scope") } }
		return handler(ctx, req)
	}
}

// unaryRateLimit enforces per-tenant+method rate limiting.
func unaryRateLimit(l *security.RateLimiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		key := "anon:" + info.FullMethod
		if c, ok := handlers.ClaimsFromContext(ctx); ok {
			key = c.TenantID + ":" + info.FullMethod
		}
		if !l.Allow(key) {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(ctx, req)
	}
}

// unaryRBACAugment augments scopes based on roles metadata (x-roles header).
func unaryRBACAugment() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		claims, ok := handlers.ClaimsFromContext(ctx)
        if !ok { return handler(ctx, req) }
        roles := make([]string, len(claims.Roles))
        copy(roles, claims.Roles)
        aug := claims.Scopes
        for _, role := range roles {
            switch strings.ToLower(strings.TrimSpace(role)) {
            case "admin":
                aug = append(aug, "secrets:write", "parity:compute", "leaks:write", "secrets:read", "parity:read", "leaks:read")
            case "auditor":
                aug = append(aug, "secrets:read", "parity:read", "leaks:read")
            }
        }
        ctx = handlers.WithClaims(ctx, handlers.Claims{Subject: claims.Subject, TenantID: claims.TenantID, ProjectID: claims.ProjectID, Scopes: dedupScopes(aug), Roles: roles})
		return handler(ctx, req)
	}
}

func dedupScopes(in []string) []string {
    m := map[string]struct{}{}
    out := make([]string, 0, len(in))
    for _, s := range in {
        if s == "" { continue }
        if _, ok := m[s]; ok { continue }
        m[s] = struct{}{}
        out = append(out, s)
    }
    return out
}

// unaryEnvAllowlist validates env parameters present in requests and rejects disallowed values.
// Validation rules:
// - GetSecret: in.Environment
// - ListSecrets: in.Environment
// - CheckEnvParity: in.BaseEnv and in.CompareEnv
func unaryEnvAllowlist(allowed []string) grpc.UnaryServerInterceptor {
    norm := func(v string) string { return strings.ToLower(strings.TrimSpace(v)) }
    isAllowed := func(v string) bool {
        v = norm(v)
        if v == "" { return true }
        for _, a := range allowed { if norm(a) == v { return true } }
        return false
    }
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        switch r := req.(type) {
        case *pb.GetSecretRequest:
            if !isAllowed(r.Environment) { return nil, status.Error(codes.InvalidArgument, "environment not allowed") }
        case *pb.ListSecretsRequest:
            if !isAllowed(r.Environment) { return nil, status.Error(codes.InvalidArgument, "environment not allowed") }
        case *pb.CheckEnvParityRequest:
            if !isAllowed(r.BaseEnv) || !isAllowed(r.CompareEnv) { return nil, status.Error(codes.InvalidArgument, "environment not allowed") }
        }
        return handler(ctx, req)
    }
}

// unaryGRPCMetrics records gRPC request count and duration by method and final status code.
func unaryGRPCMetrics(reg prometheus.Registerer) grpc.UnaryServerInterceptor {
    requests := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "grpc_requests_total",
            Help: "Total number of gRPC requests by method and code.",
        },
        []string{"method", "code"},
    )
    if err := reg.Register(requests); err != nil {
        if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
            if cv, ok := are.ExistingCollector.(*prometheus.CounterVec); ok { requests = cv }
        }
    }
    durations := prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "grpc_request_duration_seconds",
            Help:    "gRPC request duration in seconds by method and code.",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "code"},
    )
    if err := reg.Register(durations); err != nil {
        if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
            if hv, ok := are.ExistingCollector.(*prometheus.HistogramVec); ok { durations = hv }
        }
    }

    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        resp, err := handler(ctx, req)
        code := codes.OK
        if st, ok := status.FromError(err); ok { code = st.Code() }
        method := info.FullMethod
        requests.WithLabelValues(method, code.String()).Inc()
        durations.WithLabelValues(method, code.String()).Observe(time.Since(start).Seconds())
        return resp, err
    }
}
