package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapi "example.com/lma/snapshot-orchestrator-service/internal/http"
	"example.com/lma/snapshot-orchestrator-service/internal/snapshots"
)

func main() {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "7054"
	}

	store := snapshots.NewStore()
	handler := httpapi.NewRouter(store)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("snapshot-orchestrator-service listening on :%s", port)
		errCh <- srv.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		fmt.Println(err)
	}
}
