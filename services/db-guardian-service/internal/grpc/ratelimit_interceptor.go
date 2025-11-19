package grpcserver

import (
    "context"
    "net"
    "strings"

    "example.com/lma/db-guardian-service/internal/ratelimit"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/peer"
    "google.golang.org/grpc/status"
)

// RateLimitUnaryInterceptor enforces rate limits using provided limiter, key, and cost functions.
func RateLimitUnaryInterceptor(limiter ratelimit.RateLimiter, keyFn func(ctx context.Context, method string) string, costFn func(method string) int) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        if limiter == nil { // interceptor is typically not added when nil, guard anyway
            return handler(ctx, req)
        }
        key := "anonymous"
        if keyFn != nil { key = keyFn(ctx, info.FullMethod) }
        cost := 1
        if costFn != nil { cost = costFn(info.FullMethod) }
        ok, err := limiter.Allow(ctx, key, cost)
        if err != nil {
            return nil, status.Error(codes.Unavailable, "rate limit error")
        }
        if !ok {
            return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
        }
        return handler(ctx, req)
    }
}

// DefaultGRPCKeyFn derives a key from subject claims or peer IP.
func DefaultGRPCKeyFn(ctx context.Context, _ string) string {
    if c := ClaimsFromContext(ctx); c != nil && c.Subject != "" {
        return c.Subject
    }
    if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
        host, _, err := net.SplitHostPort(p.Addr.String())
        if err == nil && host != "" {
            return host
        }
        return p.Addr.String()
    }
    return "anonymous"
}

// DefaultGRPCCostFn assigns higher costs to write methods.
func DefaultGRPCCostFn(method string) int {
    if strings.Contains(method, "/RegisterConnection") ||
        strings.Contains(method, "/UpdatePolicy") ||
        strings.Contains(method, "/ValidateMigration") ||
        strings.Contains(method, "/RunAnalysis") {
        return 5
    }
    return 1
}
