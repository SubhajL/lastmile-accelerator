package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel"
)

func TestOtelTracingMiddleware_RecordsSpan(t *testing.T) {
	// Setup in-memory exporter
	exp := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exp))
	otel.SetTracerProvider(tp)

	next := OtelTracing()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	w := httptest.NewRecorder()
	next.ServeHTTP(w, r)
	if w.Code != 204 { t.Fatalf("unexpected status") }

	spans := exp.GetSpans()
	if len(spans) == 0 {
		t.Fatalf("expected spans to be recorded")
	}
}
