package handlers

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-chi/chi/v5"
    "github.com/prometheus/client_golang/prometheus"
)

func TestRoutePattern_ReturnsChiPattern(t *testing.T) {
    r := httptest.NewRequest(http.MethodGet, "/v1/projects/123/secrets", nil)
    w := httptest.NewRecorder()
    router := chi.NewRouter()
    router.Get("/v1/projects/{projectID}/secrets", func(w http.ResponseWriter, r *http.Request) {
        if got := routePattern(r); got != "/v1/projects/{projectID}/secrets" {
            t.Fatalf("route pattern mismatch: %s", got)
        }
        w.WriteHeader(http.StatusOK)
    })
    router.ServeHTTP(w, r)
    if w.Code != 200 { t.Fatalf("unexpected status: %d", w.Code) }
}

func TestHttpMetrics_IncrementsCounter_OnSuccess(t *testing.T) {
    reg := prometheus.NewRegistry()
    router := chi.NewRouter()
    router.Use(HttpMetrics(reg))
    router.Get("/v1/projects/{projectID}/secrets", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/v1/projects/abc/secrets", nil)
    router.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("unexpected status: %d", rr.Code) }

    fams, err := reg.Gather()
    if err != nil { t.Fatalf("gather: %v", err) }
    found := false
    for _, f := range fams {
        if f.GetName() == "http_requests_total" {
            for _, m := range f.GetMetric() {
                var method, route, status string
                for _, lp := range m.GetLabel() {
                    switch lp.GetName() {
                    case "method": method = lp.GetValue()
                    case "route": route = lp.GetValue()
                    case "status": status = lp.GetValue()
                    }
                }
                if method == "GET" && route == "/v1/projects/{projectID}/secrets" && status == "200" && m.GetCounter().GetValue() >= 1 {
                    found = true
                    break
                }
            }
        }
    }
    if !found { t.Fatalf("expected http_requests_total with labels and count >=1") }
}

func TestHttpMetrics_ObservesLatencyHistogram(t *testing.T) {
    reg := prometheus.NewRegistry()
    router := chi.NewRouter()
    router.Use(HttpMetrics(reg))
    router.Get("/v1/projects/{projectID}/secrets", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/v1/projects/abc/secrets", nil)
    router.ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("unexpected status: %d", rr.Code) }
    fams, err := reg.Gather()
    if err != nil { t.Fatalf("gather: %v", err) }
    found := false
    for _, f := range fams {
        if f.GetName() == "http_request_duration_seconds" {
            for _, m := range f.GetMetric() {
                if m.GetHistogram().GetSampleCount() > 0 { found = true; break }
            }
        }
    }
    if !found { t.Fatalf("expected histogram sample count > 0") }
}
