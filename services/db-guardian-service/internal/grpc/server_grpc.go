package grpcserver

import (
	"context"
	"net"

	"example.com/lma/db-guardian-service/dbguardian/v1"
	"example.com/lma/db-guardian-service/internal/server"
	grpc "google.golang.org/grpc"
)

// NewGRPCServer creates a gRPC server with registered services.
func NewGRPCServer(deps *server.Dependencies, opts ...grpc.ServerOption) *grpc.Server {
	// default interceptors: tracing + logging; auth if provided
	var intrs []grpc.UnaryServerInterceptor
	intrs = append(intrs, TracingUnaryInterceptor())
    if deps != nil && deps.Logger != nil { intrs = append(intrs, LoggingUnaryInterceptor(deps.Logger)) }
    if deps != nil && deps.Authenticator != nil { intrs = append(intrs, AuthUnaryInterceptor(deps.Authenticator, server.GRPCScopeResolver)) }

	if len(intrs) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(intrs...))
	}

	s := grpc.NewServer(opts...)
	dbguardianv1.RegisterDbGuardianServiceServer(s, newGuardianServer(deps))
	return s
}

// StartGRPC starts serving on the given listener until context is canceled.
func StartGRPC(ctx context.Context, lis net.Listener, s *grpc.Server) error {
	errCh := make(chan error, 1)
	go func() { errCh <- s.Serve(lis) }()
	select {
	case <-ctx.Done():
		s.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}
