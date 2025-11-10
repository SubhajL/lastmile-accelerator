package main

import (
	"context"
	"net"

	grpcserver "example.com/lma/db-guardian-service/internal/grpc"
	"example.com/lma/db-guardian-service/internal/server"
	"example.com/lma/db-guardian-service/pkg/logger"
)

func maybeStartGRPC(ctx context.Context, deps *server.Dependencies, log *logger.Logger) (func(), error) {
	lis, err := net.Listen("tcp", ":50065")
	if err != nil { return nil, err }
	s := grpcserver.NewGRPCServer(deps)
	go func() { _ = grpcserver.StartGRPC(ctx, lis, s) }()
	if log != nil { log.Info("gRPC server started", logger.Field{Key: "port", Value: 50065}) }
	return func() { s.GracefulStop(); _ = lis.Close() }, nil
}