package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestNew_CreatesLoggerWithServiceName(t *testing.T) {
	// Arrange
	var buf bytes.Buffer

	// Act
	log := New("test-service", "dev", &buf)

	// Assert
	if log == nil {
		t.Fatal("expected logger instance, got nil")
	}
	
	log.Info("test message")
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}
	
	if logEntry["service.name"] != "test-service" {
		t.Errorf("expected service.name 'test-service', got '%v'", logEntry["service.name"])
	}
}

func TestWithContext_WithTraceID_IncludesTraceInLog(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	log := New("test-service", "dev", &buf)
	
	// Create a span context with trace ID
	traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanCtx)

	// Act
	logWithCtx := log.WithContext(ctx)
	logWithCtx.Info("test with trace")

	// Assert
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}
	
	if logEntry["trace_id"] == nil {
		t.Error("expected trace_id in log, got nil")
	}
}

func TestInfo_LogsAtCorrectLevel(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	log := New("test-service", "dev", &buf)

	// Act
	log.Info("info message")

	// Assert
	output := buf.String()
	if !strings.Contains(output, "info message") {
		t.Errorf("expected log to contain 'info message', got '%s'", output)
	}
	
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}
	
	if logEntry["level"] != "info" {
		t.Errorf("expected level 'info', got '%v'", logEntry["level"])
	}
}

func TestError_WithFields_IncludesAllFields(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	log := New("test-service", "dev", &buf)

	// Act
	log.Error("error occurred", 
		Field{Key: "error_code", Value: 500},
		Field{Key: "user_id", Value: "user123"},
	)

	// Assert
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}
	
	if logEntry["level"] != "error" {
		t.Errorf("expected level 'error', got '%v'", logEntry["level"])
	}
	if logEntry["error_code"] != float64(500) {
		t.Errorf("expected error_code 500, got '%v'", logEntry["error_code"])
	}
	if logEntry["user_id"] != "user123" {
		t.Errorf("expected user_id 'user123', got '%v'", logEntry["user_id"])
	}
}
