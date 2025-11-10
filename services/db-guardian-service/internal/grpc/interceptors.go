package grpcserver

import (
	"context"
	"time"

	"example.com/lma/db-guardian-service/internal/auth"
	"example.com/lma/db-guardian-service/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"example.com/lma/db-guardian-service/internal/telemetry"
)

// context key for claims
type claimsKeyType struct{}
var claimsKey claimsKeyType

// ClaimsFromContext extracts auth claims if set
func ClaimsFromContext(ctx context.Context) *auth.Claims {
	v := ctx.Value(claimsKey)
	if v == nil { return nil }
	if c, ok := v.(*auth.Claims); ok { return c }
	return nil
}

// AuthUnaryInterceptor enforces JWT scopes via provided authenticator and resolver.
func AuthUnaryInterceptor(authn auth.Authenticator, resolve func(method string) []string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if authn == nil { return handler(ctx, req) }
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok { return nil, status.Error(codes.Unauthenticated, "missing metadata") }
		authz := md.Get("authorization")
		if len(authz) == 0 || authz[0] == "" { return nil, status.Error(codes.Unauthenticated, "missing authorization") }
		required := []string{}
		if resolve != nil { required = resolve(info.FullMethod) }
		claims, err := authn.Verify(ctx, authz[0], required)
		if err != nil { return nil, status.Error(codes.PermissionDenied, "forbidden") }
		ctx = context.WithValue(ctx, claimsKey, claims)
		return handler(ctx, req)
	}
}

// LoggingUnaryInterceptor logs method, code, and duration.
func LoggingUnaryInterceptor(l *logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err)
		if l != nil {
			l.WithContext(ctx).Info("grpc request",
				logger.Field{Key: "grpc.method", Value: info.FullMethod},
				logger.Field{Key: "grpc.code", Value: code.String()},
				logger.Field{Key: "duration_ms", Value: time.Since(start).Milliseconds()},
			)
		}
		return resp, err
	}
}

// TracingUnaryInterceptor starts a server span per request.
func TracingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, span := telemetry.StartSpan(ctx, info.FullMethod,
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
			),
		)
		defer span.End()
		return handler(ctx, req)
	}
}