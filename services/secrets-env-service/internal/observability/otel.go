package observability

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"example.com/lma/secrets-env-service/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.24.0"
)

func normalizeOTLPEndpoint(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	if strings.Contains(raw, "://") {
		parsed, err := url.Parse(raw)
		if err != nil {
			return "", fmt.Errorf("invalid OTEL_EXPORTER_OTLP_ENDPOINT: %w", err)
		}
		if parsed.Host == "" {
			return "", fmt.Errorf("invalid OTEL_EXPORTER_OTLP_ENDPOINT: missing host")
		}
		if parsed.Path != "" && parsed.Path != "/" {
			return "", fmt.Errorf("invalid OTEL_EXPORTER_OTLP_ENDPOINT: path not supported for gRPC exporter")
		}
		return parsed.Host, nil
	}

	if strings.Contains(raw, "/") {
		return "", fmt.Errorf("invalid OTEL_EXPORTER_OTLP_ENDPOINT: path not supported for gRPC exporter")
	}

	return raw, nil
}

// Init sets a global tracer provider.
func Init(ctx context.Context, cfg config.ObservabilityConfig, serviceName string) (func(context.Context) error, error) {
	resolvedServiceName := serviceName
	if cfg.OTELServiceName != "" {
		resolvedServiceName = cfg.OTELServiceName
	}

	res, _ := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(resolvedServiceName)))

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
	}

	if cfg.OTELEndpoint != "" {
		endpoint, err := normalizeOTLPEndpoint(cfg.OTELEndpoint)
		if err != nil {
			return nil, err
		}

		expOpts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(endpoint),
		}
		if cfg.OTELInsecure {
			expOpts = append(expOpts, otlptracegrpc.WithInsecure())
		}
		if len(cfg.OTELHeaders) > 0 {
			expOpts = append(expOpts, otlptracegrpc.WithHeaders(cfg.OTELHeaders))
		}

		exporter, err := otlptracegrpc.New(ctx, expOpts...)
		if err != nil {
			return nil, err
		}

		opts = append(opts, sdktrace.WithBatcher(exporter))
	}

	tp := sdktrace.NewTracerProvider(opts...)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
