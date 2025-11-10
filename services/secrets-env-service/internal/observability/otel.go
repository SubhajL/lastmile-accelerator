package observability

import (
	"context"

	"example.com/lma/secrets-env-service/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.24.0"
)

// Init sets a global tracer provider. This minimal setup creates a provider with a resource; exporting is left to the environment/auto-instrumentation.
func Init(ctx context.Context, cfg config.ObservabilityConfig, serviceName string) (func(context.Context) error, error) {
	res, _ := resource.New(ctx, resource.WithAttributes(semconv.ServiceName(serviceName)))
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}
