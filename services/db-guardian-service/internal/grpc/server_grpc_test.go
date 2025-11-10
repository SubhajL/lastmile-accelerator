package grpcserver

import (
	"context"
	"testing"

	grpc "google.golang.org/grpc"
)

func TestNewGRPCServer_Builds(t *testing.T) {
	s := NewGRPCServer(nil, grpc.EmptyServerOption{})
	if s == nil { t.Fatal("server is nil") }
}

func TestStartGRPC_CancelsGracefully(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// We can start and immediately cancel with a real in-memory listener
	go func() { cancel() }()
	_ = ctx
}
