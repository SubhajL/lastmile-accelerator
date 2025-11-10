package telemetry

import (
	"context"
	"fmt"

	"example.com/lma/db-guardian-service/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

func InitTracer(ctx context.Context, cfg *config.Config) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion("1.0.0"),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var tracerProvider *sdktrace.TracerProvider

	if cfg.OTLPEndpoint != "" {
		exporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(extractHost(cfg.OTLPEndpoint)),
			otlptracehttp.WithInsecure(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
		}

		tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
		)
	} else {
		// No exporter configured, use basic provider
		tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
		)
	}

	otel.SetTracerProvider(tracerProvider)

	shutdown := func(ctx context.Context) error {
		return tracerProvider.Shutdown(ctx)
	}

	return shutdown, nil
}

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := otel.Tracer("db-guardian-service")
	return tracer.Start(ctx, name, opts...)
}

func extractHost(endpoint string) string {
	// Remove http:// or https:// prefix
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		return endpoint[7:]
	}
	if len(endpoint) > 8 && endpoint[:8] == "https://" {
		return endpoint[8:]
	}
	return endpoint
}
