package main

import (
    "context"
    "fmt"

    "example.com/lma/secrets-env-service/internal/config"
    "example.com/lma/secrets-env-service/internal/observability"
)

// initOTel validates the observability config and initializes OTEL exporter.
func initOTel(ctx context.Context, oc config.ObservabilityConfig, defaultServiceName string) (func(context.Context) error, error) {
    if oc.OTELEndpoint == "" {
        return nil, fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT is required")
    }
    return observability.Init(ctx, oc, defaultServiceName)
}
