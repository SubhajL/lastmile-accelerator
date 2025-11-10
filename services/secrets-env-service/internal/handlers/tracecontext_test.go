package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/lma/secrets-env-service/internal/events"
)

func TestTraceContext_UsesHeaderOrGenerates(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ensure downstream can read traceparent from context
		tp := events.TraceparentFromContext(r.Context())
		if tp == "" { t.Fatalf("missing traceparent") }
		w.WriteHeader(204)
	})
	rec := TraceContext()(next)

	// with header
	r := httptest.NewRequest(http.MethodGet, "/x", nil)
	r.Header.Set("traceparent", "00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01")
	w := httptest.NewRecorder()
	rec.ServeHTTP(w, r)
	if w.Code != 204 { t.Fatalf("unexpected status") }

	// without header (should still set something)
	r2 := httptest.NewRequest(http.MethodGet, "/x", nil)
	w2 := httptest.NewRecorder()
	rec.ServeHTTP(w2, r2)
	if w2.Code != 204 { t.Fatalf("unexpected status") }
}
