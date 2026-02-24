package grpcapi

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"example.com/lma/secrets-env-service/internal/domain"
	"example.com/lma/secrets-env-service/internal/events"
	pb "example.com/lma/secrets-env-service/internal/grpc/gen/secretsenvv1"
	"example.com/lma/secrets-env-service/internal/handlers"
	"example.com/lma/secrets-env-service/internal/logger"
	"example.com/lma/secrets-env-service/internal/security"
	"example.com/lma/secrets-env-service/internal/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
	Meta  SecretMeta     `json:"meta"`
	Value map[string]any `json:"value"`
}

type ListSecretsRequest struct {
	ProjectID   string `json:"project_id"`
	Environment string `json:"environment"`
	PageSize    int32  `json:"page_size"`
	PageToken   string `json:"page_token"`
}

type ListSecretsResponse struct {
	Items         []SecretMeta `json:"items"`
	NextPageToken string       `json:"next_page_token"`
}

type CheckEnvParityRequest struct {
	ProjectID  string `json:"project_id"`
	BaseEnv    string `json:"base_env"`
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
func StartGRPCServer(ctx context.Context, addr string, secrets *service.SecretsService, parity *service.ParityService, verifier handlers.TokenVerifier, log zerolog.Logger, tlsConfig *tls.Config) error {
	// Interceptor chain: recovery -> logging -> auth -> scope enforcement
	svc := &server{secrets: secrets, parity: parity, log: log}

	requiredScopes := map[string][]string{
		"/lma.secretsenv.v1.SecretsEnvService/GetSecret":      {"secrets:read"},
		"/lma.secretsenv.v1.SecretsEnvService/ListSecrets":    {"secrets:read"},
		"/lma.secretsenv.v1.SecretsEnvService/CheckEnvParity": {"parity:compute"},
	}

	chain := grpc.ChainUnaryInterceptor(
		unaryPanicRecovery(log),
		unaryGRPCMetrics(prometheus.DefaultRegisterer),
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

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		gs.GracefulStop()
		_ = lis.Close()
	}()
	return gs.Serve(lis)
}

// Handlers (pb interfaces)

func (s *server) GetSecret(ctx context.Context, in *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	claims, ok := handlers.ClaimsFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing claims")
	}
	sec, val, err := s.secrets.GetSecret(ctx, claims.TenantID, in.ProjectId, in.Key, in.Environment)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil, status.Error(codes.NotFound, "not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.GetSecretResponse{Meta: toPBSecretMeta(sec), Value: val}, nil
}

func (s *server) ListSecrets(ctx context.Context, in *pb.ListSecretsRequest) (*pb.ListSecretsResponse, error) {
	limit := int(in.PageSize)
	if limit <= 0 {
		limit = 50
	}
	items, next, err := s.secrets.ListSecrets(ctx, in.ProjectId, in.Environment, limit, in.PageToken)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*pb.SecretMeta, 0, len(items))
	for _, it := range items {
		out = append(out, toPBSecretMeta(it))
	}
	return &pb.ListSecretsResponse{Items: out, NextPageToken: next}, nil
}

func (s *server) CheckEnvParity(ctx context.Context, in *pb.CheckEnvParityRequest) (*pb.CheckEnvParityResponse, error) {
	res, err := s.parity.CheckParity(ctx, in.ProjectId, in.BaseEnv, in.CompareEnv)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.CheckEnvParityResponse{
		ProjectId:   res.ProjectID,
		MissingKeys: res.MissingKeys,
		ExtraKeys:   res.ExtraKeys,
		HasDrift:    res.HasDrift(),
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
			if len(vals) > 0 {
				tp = vals[0]
			}
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
		if st, ok := status.FromError(err); ok {
			code = st.Code()
		}
		log.Info().Str("method", info.FullMethod).Str("grpc_code", code.String()).Dur("duration_ms", dur).Msg("grpc")
		return resp, err
	}
}

func unaryGRPCMetrics(reg prometheus.Registerer) grpc.UnaryServerInterceptor {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}

	requests := registerOrGetCounterVec(reg, prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests by method and status.",
		},
		[]string{"method", "status"},
	))
	durations := registerOrGetHistogramVec(reg, prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration by method.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	))

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)

		code := status.Code(err).String()
		requests.WithLabelValues(info.FullMethod, code).Inc()
		durations.WithLabelValues(info.FullMethod).Observe(time.Since(start).Seconds())

		return resp, err
	}
}

func registerOrGetCounterVec(reg prometheus.Registerer, vec *prometheus.CounterVec) *prometheus.CounterVec {
	if err := reg.Register(vec); err != nil {
		var already prometheus.AlreadyRegisteredError
		if errors.As(err, &already) {
			if existing, ok := already.ExistingCollector.(*prometheus.CounterVec); ok {
				return existing
			}
			return vec
		}
		panic(fmt.Errorf("register counter vec: %w", err))
	}
	return vec
}

func registerOrGetHistogramVec(reg prometheus.Registerer, vec *prometheus.HistogramVec) *prometheus.HistogramVec {
	if err := reg.Register(vec); err != nil {
		var already prometheus.AlreadyRegisteredError
		if errors.As(err, &already) {
			if existing, ok := already.ExistingCollector.(*prometheus.HistogramVec); ok {
				return existing
			}
			return vec
		}
		panic(fmt.Errorf("register histogram vec: %w", err))
	}
	return vec
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
		if err != nil || claims == nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		ctx = handlers.WithClaims(ctx, *claims)
		return handler(ctx, req)
	}
}

func unaryRequireScopes(required map[string][]string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		reqScopes := required[info.FullMethod]
		if len(reqScopes) == 0 {
			return handler(ctx, req)
		}
		claims, ok := handlers.ClaimsFromContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "missing claims")
		}
		set := map[string]struct{}{}
		for _, s := range claims.Scopes {
			set[s] = struct{}{}
		}
		for _, need := range reqScopes {
			if _, ok := set[need]; !ok {
				return nil, status.Error(codes.PermissionDenied, "insufficient scope")
			}
		}
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
		if !ok {
			return handler(ctx, req)
		}
		md, _ := metadata.FromIncomingContext(ctx)
		var roles []string
		if md != nil {
			roles = md.Get("x-roles")
		}
		aug := claims.Scopes
		for _, role := range roles {
			switch strings.ToLower(strings.TrimSpace(role)) {
			case "admin":
				aug = append(aug, "secrets:write", "parity:compute", "leaks:write")
			case "auditor":
				aug = append(aug, "secrets:read", "parity:read", "leaks:read")
			}
		}
		ctx = handlers.WithClaims(ctx, handlers.Claims{Subject: claims.Subject, TenantID: claims.TenantID, ProjectID: claims.ProjectID, Scopes: aug})
		return handler(ctx, req)
	}
}
