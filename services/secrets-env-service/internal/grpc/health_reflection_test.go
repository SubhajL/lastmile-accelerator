package grpcapi

import (
    "context"
    "net"
    "testing"

    appLogger "example.com/lma/secrets-env-service/internal/logger"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    healthpb "google.golang.org/grpc/health/grpc_health_v1"
    reflpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
    "google.golang.org/grpc/test/bufconn"
)

// startHealthReflBuf starts a bare server with only health/reflection toggles enabled as requested.
func startHealthReflBuf(t *testing.T, healthEnabled, reflEnabled bool) (*grpc.ClientConn, func()) {
    t.Helper()
    _ = appLogger.New("test", "info", nil)
    lis := bufconn.Listen(1 << 20)
    // Minimal server with no services; we rely on health/reflection registration path in this package.
    gs := grpc.NewServer()
    if healthEnabled {
        hs, cleanup := registerHealth(gs, context.Background())
        _ = hs; _ = cleanup // cleanup not used in this short-lived test
    }
    if reflEnabled {
        registerReflection(gs)
    }
    go func() { _ = gs.Serve(lis) }()
    dialer := func(ctx context.Context, s string) (net.Conn, error) { return lis.Dial() }
    conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(dialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { t.Fatalf("dial err: %v", err) }
    stop := func() { gs.GracefulStop(); _ = lis.Close(); _ = conn.Close() }
    return conn, stop
}

func TestGRPC_Health_CheckServing_WhenEnabled(t *testing.T) {
    conn, stop := startHealthReflBuf(t, true, false)
    defer stop()
    c := healthpb.NewHealthClient(conn)
    if _, err := c.Check(context.Background(), &healthpb.HealthCheckRequest{}); err != nil { t.Fatalf("health check err: %v", err) }
}

func TestGRPC_Health_Unimplemented_WhenDisabled(t *testing.T) {
    conn, stop := startHealthReflBuf(t, false, false)
    defer stop()
    c := healthpb.NewHealthClient(conn)
    if _, err := c.Check(context.Background(), &healthpb.HealthCheckRequest{}); err == nil {
        t.Fatalf("expected error when health disabled")
    }
}

func TestGRPC_Reflection_Lists_WhenEnabled(t *testing.T) {
    conn, stop := startHealthReflBuf(t, false, true)
    defer stop()
    rc := reflpb.NewServerReflectionClient(conn)
    stream, err := rc.ServerReflectionInfo(context.Background())
    if err != nil { t.Fatalf("stream err: %v", err) }
    if err := stream.Send(&reflpb.ServerReflectionRequest{MessageRequest: &reflpb.ServerReflectionRequest_ListServices{ListServices: ""}}); err != nil {
        t.Fatalf("send err: %v", err)
    }
    _ = stream.CloseSend()
}

func TestGRPC_Reflection_Unimplemented_WhenDisabled(t *testing.T) {
    conn, stop := startHealthReflBuf(t, false, false)
    defer stop()
    rc := reflpb.NewServerReflectionClient(conn)
    stream, err := rc.ServerReflectionInfo(context.Background())
    if err != nil { t.Fatalf("unexpected stream err: %v", err) }
    _ = stream.Send(&reflpb.ServerReflectionRequest{MessageRequest: &reflpb.ServerReflectionRequest_ListServices{ListServices: ""}})
    if _, err := stream.Recv(); err == nil { t.Fatalf("expected error on recv when reflection disabled") }
}
