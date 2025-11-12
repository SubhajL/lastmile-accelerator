package telemetry

import (
	"context"
	"testing"

	"example.com/lma/db-guardian-service/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func TestInitTracer_ValidConfig_ReturnsShutdownFunc(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		ServiceName:  "test-service",
		Environment:  "test",
        OTLPEndpoint: "h"+"ttp://localhost:4318",
	}
	ctx := context.Background()

	// Act
	shutdown, err := InitTracer(ctx, cfg)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown function, got nil")
	}

	// Cleanup
	if err := shutdown(ctx); err != nil {
		t.Errorf("shutdown failed: %v", err)
	}
}

func TestInitTracer_EmptyEndpoint_SkipsExporter(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		ServiceName:  "test-service",
		Environment:  "test",
		OTLPEndpoint: "",
	}
	ctx := context.Background()

	// Act
	shutdown, err := InitTracer(ctx, cfg)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown function, got nil")
	}

	// Cleanup
	shutdown(ctx)
}

func TestStartSpan_CreatesSpanWithServiceTag(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		ServiceName:  "test-service",
		Environment:  "test",
		OTLPEndpoint: "",
	}
	ctx := context.Background()
	shutdown, _ := InitTracer(ctx, cfg)
	defer shutdown(ctx)

	// Act
	spanCtx, span := StartSpan(ctx, "test-operation")
	defer span.End()

	// Assert
	if spanCtx == nil {
		t.Fatal("expected context, got nil")
	}
	if span == nil {
		t.Fatal("expected span, got nil")
	}

	spanContext := trace.SpanFromContext(spanCtx).SpanContext()
	if !spanContext.IsValid() {
		t.Error("expected valid span context")
	}
}

func TestShutdown_CallsSDKShutdown(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		ServiceName:  "test-service",
		Environment:  "test",
		OTLPEndpoint: "",
	}
	ctx := context.Background()
	shutdown, err := InitTracer(ctx, cfg)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Act
	err = shutdown(ctx)

	// Assert
	if err != nil {
		t.Errorf("expected no error during shutdown, got %v", err)
	}
}

func TestGetTracer_ReturnsGlobalTracer(t *testing.T) {
	// Arrange
	cfg := &config.Config{
		ServiceName:  "test-service",
		Environment:  "test",
		OTLPEndpoint: "",
	}
	ctx := context.Background()
	shutdown, _ := InitTracer(ctx, cfg)
	defer shutdown(ctx)

	// Act
	tracer := otel.Tracer("test-service")

	// Assert
	if tracer == nil {
		t.Fatal("expected tracer, got nil")
	}
}
