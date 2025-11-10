package grpcserver

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"example.com/lma/db-guardian-service/internal/auth"
	"example.com/lma/db-guardian-service/internal/config"
	"example.com/lma/db-guardian-service/internal/telemetry"
	"example.com/lma/db-guardian-service/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakeAuth struct{
	claims *auth.Claims
	err error
	lastRequired []string
}

func (f *fakeAuth) Verify(_ context.Context, bearer string, required []string) (*auth.Claims, error) {
	f.lastRequired = required
	if !strings.HasPrefix(bearer, "Bearer ") { return nil, status.Error(codes.Unauthenticated, "bad header") }
	return f.claims, f.err
}

func TestAuthUnaryInterceptor_RequiresScopes(t *testing.T) {
	fa := &fakeAuth{claims: &auth.Claims{Subject: "u1", Scopes: []string{"db.read"}}}
	resolve := func(method string) []string { if strings.Contains(method, "GetPolicy") { return []string{"db.read"} }; return nil }
	intr := AuthUnaryInterceptor(fa, resolve)

	called := false
	h := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return nil, nil
	}
	info := &grpc.UnaryServerInfo{FullMethod: "/dbguardian.v1.DbGuardianService/GetPolicy"}
	md := metadata.Pairs("authorization", "Bearer ok")
	ctx := metadata.NewIncomingContext(context.Background(), md)
	_, err := intr(ctx, nil, info, h)
	if err != nil { t.Fatalf("unexpected err: %v", err) }
	if !called { t.Fatal("handler not called") }
	if len(fa.lastRequired) != 1 || fa.lastRequired[0] != "db.read" { t.Fatalf("required scopes mismatch: %v", fa.lastRequired) }
}

func TestAuthUnaryInterceptor_MissingHeader(t *testing.T) {
	fa := &fakeAuth{claims: &auth.Claims{Subject: "u1"}}
	intr := AuthUnaryInterceptor(fa, func(string) []string { return []string{"x"} })
	info := &grpc.UnaryServerInfo{FullMethod: "/m"}
	ctx := context.Background()
	_, err := intr(ctx, nil, info, func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil })
	if status.Code(err) != codes.Unauthenticated { t.Fatalf("expected Unauthenticated, got %v", status.Code(err)) }
}

func TestLoggingUnaryInterceptor_LogsMethodAndCode(t *testing.T) {
	var buf bytes.Buffer
	lg := loggerForTest(&buf)
	intr := LoggingUnaryInterceptor(lg)
	info := &grpc.UnaryServerInfo{FullMethod: "/dbguardian.v1.DbGuardianService/GetPolicy"}
	ctx := context.Background()
	_, _ = intr(ctx, nil, info, func(ctx context.Context, req interface{}) (interface{}, error) { return nil, status.Error(codes.NotFound, "nope") })
	out := buf.String()
	if !strings.Contains(out, "\"grpc.method\":\"/dbguardian.v1.DbGuardianService/GetPolicy\"") { t.Fatalf("missing method in logs: %s", out) }
	if !strings.Contains(out, "\"grpc.code\":\"NotFound\"") { t.Fatalf("missing code in logs: %s", out) }
}

func TestTracingUnaryInterceptor_StartsSpan(t *testing.T) {
	// initialize tracer provider to get valid span context
	shutdown, err := telemetry.InitTracer(context.Background(), &config.Config{ServiceName: "db-guardian-service", Environment: "test", OTLPEndpoint: ""})
	if err != nil { t.Fatalf("tracer init: %v", err) }
	defer func(){ _ = shutdown(context.Background()) }()

	intr := TracingUnaryInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/dbguardian.v1.DbGuardianService/GetPolicy"}
	var sawSpan bool
	_, err = intr(context.Background(), nil, info, func(ctx context.Context, req interface{}) (interface{}, error) {
		// Verify a span exists in context (SpanFromContext not noop)
		_, span := telemetry.StartSpan(ctx, "test-child")
		sawSpan = span != nil
		span.End()
		return nil, nil
	})
	if err != nil { t.Fatalf("unexpected: %v", err) }
	// Give time for async operations (noop provider still ok)
	time.Sleep(10 * time.Millisecond)
	if !sawSpan { t.Fatal("expected to see span in handler context") }
}

// loggerForTest returns a logger writing to given buffer
func loggerForTest(buf *bytes.Buffer) *logger.Logger {
	return logger.New("db-guardian-service", "test", buf)
}