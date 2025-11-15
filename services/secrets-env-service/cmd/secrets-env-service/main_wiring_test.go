package main

import (
    "context"
    "testing"

    "example.com/lma/secrets-env-service/internal/config"
)

func Test_initOTel_Error_WhenEndpointMissing(t *testing.T) {
    ctx := context.Background()
    oc := config.ObservabilityConfig{}
    shutdown, err := initOTel(ctx, oc, "secrets-env-service")
    if err == nil || shutdown != nil {
        t.Fatalf("expected error and nil shutdown when endpoint missing")
    }
}

func Test_initOTel_ReturnsShutdown_WhenEndpointSet(t *testing.T) {
    ctx := context.Background()
    oc := config.ObservabilityConfig{OTELEndpoint: "localhost:4317", OTELInsecure: true}
    shutdown, err := initOTel(ctx, oc, "secrets-env-service")
    if err != nil || shutdown == nil {
        t.Fatalf("expected success and non-nil shutdown")
    }
}
