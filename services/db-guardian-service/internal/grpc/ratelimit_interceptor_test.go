package grpcserver

import (
    "context"
    "net"
    "testing"

    "example.com/lma/db-guardian-service/internal/auth"
    "example.com/lma/db-guardian-service/internal/ratelimit"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/peer"
    "google.golang.org/grpc/status"
)

type fakeLimiter struct{
    ok bool
    err error
    lastKey string
    lastCost int
}

func (f *fakeLimiter) Allow(_ context.Context, key string, cost int) (bool, error) {
    f.lastKey = key
    f.lastCost = cost
    return f.ok, f.err
}

func TestRateLimitUnaryInterceptor_Allow_PassesThrough(t *testing.T) {
    fl := &fakeLimiter{ok: true}
    intr := RateLimitUnaryInterceptor(fl, DefaultGRPCKeyFn, func(string) int { return 3 })
    called := false
    h := func(ctx context.Context, req interface{}) (interface{}, error) { called = true; return nil, nil }
    info := &grpc.UnaryServerInfo{FullMethod: "/dbguardian.v1.DbGuardianService/GetPolicy"}
    // subject present → key is subject
    ctx := context.WithValue(context.Background(), claimsKey, &auth.Claims{Subject: "u1"})
    _, err := intr(ctx, nil, info, h)
    if err != nil { t.Fatalf("unexpected err: %v", err) }
    if !called { t.Fatal("handler not called") }
    if fl.lastKey != "u1" || fl.lastCost != 3 { t.Fatalf("limiter args mismatch: key=%s cost=%d", fl.lastKey, fl.lastCost) }
}

func TestRateLimitUnaryInterceptor_Deny_ReturnsResourceExhausted(t *testing.T) {
    fl := &fakeLimiter{ok: false}
    intr := RateLimitUnaryInterceptor(fl, DefaultGRPCKeyFn, func(string) int { return 1 })
    h := func(ctx context.Context, req interface{}) (interface{}, error) { t.Fatal("handler should not run"); return nil, nil }
    info := &grpc.UnaryServerInfo{FullMethod: "/m"}
    ctx := context.Background()
    _, err := intr(ctx, nil, info, h)
    if status.Code(err) != codes.ResourceExhausted { t.Fatalf("expected ResourceExhausted, got %v", status.Code(err)) }
}

func TestRateLimitUnaryInterceptor_Error_ReturnsUnavailable(t *testing.T) {
    fl := &fakeLimiter{ok: false, err: status.Error(codes.Internal, "boom")}
    intr := RateLimitUnaryInterceptor(fl, DefaultGRPCKeyFn, func(string) int { return 1 })
    h := func(ctx context.Context, req interface{}) (interface{}, error) { t.Fatal("handler should not run"); return nil, nil }
    info := &grpc.UnaryServerInfo{FullMethod: "/m"}
    _, err := intr(context.Background(), nil, info, h)
    if status.Code(err) != codes.Unavailable { t.Fatalf("expected Unavailable, got %v", status.Code(err)) }
}

func TestDefaultGRPCKeyFn_SubjectPreferredOverIP(t *testing.T) {
    ctx := context.WithValue(context.Background(), claimsKey, &auth.Claims{Subject: "subj"})
    if got := DefaultGRPCKeyFn(ctx, "/m"); got != "subj" { t.Fatalf("want subj, got %s", got) }
}

func TestDefaultGRPCKeyFn_FallsBackToPeerIP(t *testing.T) {
    p := &peer.Peer{Addr: &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1234}}
    ctx := peer.NewContext(context.Background(), p)
    got := DefaultGRPCKeyFn(ctx, "/m")
    if got != "1.2.3.4" { t.Fatalf("want 1.2.3.4, got %s", got) }
}

func TestDefaultGRPCCostFn_WritesHigherThanReads(t *testing.T) {
    // reads
    if c := DefaultGRPCCostFn("/dbguardian.v1.DbGuardianService/GetPolicy"); c != 1 { t.Fatalf("read cost=1, got %d", c) }
    // writes
    if c := DefaultGRPCCostFn("/dbguardian.v1.DbGuardianService/RunAnalysis"); c <= 1 { t.Fatalf("write cost>1, got %d", c) }
}

// compile-time assertion that fakeLimiter satisfies interface
var _ ratelimit.RateLimiter = (*fakeLimiter)(nil)
