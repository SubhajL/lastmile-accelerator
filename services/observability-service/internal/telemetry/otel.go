package telemetry

import (
	"context"
	"fmt"

	"example.com/lma/observability-service/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryProvider manages OTel tracer, meter, and shutdown lifecycle
type TelemetryProvider struct {
	tracerProvider *sdktrace.TracerProvider
	tracer         trace.Tracer
	meter          metric.Meter
}

// InitTelemetry initializes OTLP exporter and sets up tracer/meter providers
func InitTelemetry(cfg *config.Config) (*TelemetryProvider, error) {
	// Create resource with service name
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP trace exporter
	// Extract endpoint without scheme for otlptracehttp
	endpoint := cfg.OTelEndpoint
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		endpoint = endpoint[7:]
	} else if len(endpoint) > 8 && endpoint[:8] == "https://" {
		endpoint = endpoint[8:]
	}

	traceExporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(), // For local dev; use TLS in production
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create tracer provider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tracerProvider)

	// Get tracer
	tracer := tracerProvider.Tracer(cfg.ServiceName)

	// Get meter (using global meter provider for now)
	meter := otel.Meter(cfg.ServiceName)

	return &TelemetryProvider{
		tracerProvider: tracerProvider,
		tracer:         tracer,
		meter:          meter,
	}, nil
}

// Tracer returns configured tracer for manual span creation
func (t *TelemetryProvider) Tracer() trace.Tracer {
	return t.tracer
}

// Meter returns configured meter for custom metrics
func (t *TelemetryProvider) Meter() metric.Meter {
	return t.meter
}

// Shutdown flushes pending telemetry data and shuts down providers
func (t *TelemetryProvider) Shutdown(ctx context.Context) error {
	if t.tracerProvider != nil {
		if err := t.tracerProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown tracer provider: %w", err)
		}
	}
	return nil
}
