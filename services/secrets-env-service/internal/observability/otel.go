package observability

import (
	"context"
    "fmt"

	"example.com/lma/secrets-env-service/internal/config"
	"go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.24.0"
)

type otlpClientConfig struct {
    Endpoint string
    Insecure bool
    Headers  map[string]string
}

func buildOTLPClientConfig(cfg config.ObservabilityConfig) otlpClientConfig {
    return otlpClientConfig{
        Endpoint: cfg.OTELEndpoint,
        Insecure: cfg.OTELInsecure,
        Headers:  cfg.OTELHeaders,
    }
}

func buildResource(ctx context.Context, defaultName, override string) *resource.Resource {
    name := defaultName
    if override != "" {
        name = override
    }
    res, _ := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(name)))
    return res
}

// Init configures an OTLP gRPC exporter-backed TracerProvider and sets it globally.
func Init(ctx context.Context, cfg config.ObservabilityConfig, serviceName string) (func(context.Context) error, error) {
    ccfg := buildOTLPClientConfig(cfg)
    if ccfg.Endpoint == "" {
        return nil, fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT is required")
    }

    opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(ccfg.Endpoint)}
    if ccfg.Insecure {
        opts = append(opts, otlptracegrpc.WithInsecure())
    }
    if len(ccfg.Headers) > 0 {
        opts = append(opts, otlptracegrpc.WithHeaders(ccfg.Headers))
    }

    exp, err := otlptracegrpc.New(ctx, opts...)
    if err != nil {
        return nil, err
    }

    res := buildResource(ctx, serviceName, cfg.OTELServiceName)
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exp),
        sdktrace.WithResource(res),
    )
    otel.SetTracerProvider(tp)
    return tp.Shutdown, nil
}
