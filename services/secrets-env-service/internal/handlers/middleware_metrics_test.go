package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoutePattern_ReturnsChiPattern(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/projects/123/secrets", nil)
	w := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Get("/v1/projects/{projectID}/secrets", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/projects/{projectID}/secrets", routePattern(r))
		w.WriteHeader(http.StatusOK)
	})

	router.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
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
	assert.Equal(t, http.StatusOK, rr.Code)

	fams, err := reg.Gather()
	require.NoError(t, err)

	found := false
	for _, f := range fams {
		if f.GetName() != "http_requests_total" {
			continue
		}
		for _, m := range f.GetMetric() {
			labels := map[string]string{}
			for _, lp := range m.GetLabel() {
				labels[lp.GetName()] = lp.GetValue()
			}

			if labels["method"] == http.MethodGet &&
				labels["route"] == "/v1/projects/{projectID}/secrets" &&
				labels["status"] == "200" &&
				m.GetCounter().GetValue() >= 1 {
				found = true
				break
			}
		}
	}

	assert.True(t, found, "expected http_requests_total with labels and count >= 1")
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
	assert.Equal(t, http.StatusOK, rr.Code)

	fams, err := reg.Gather()
	require.NoError(t, err)

	found := false
	for _, f := range fams {
		if f.GetName() != "http_request_duration_seconds" {
			continue
		}
		for _, m := range f.GetMetric() {
			if m.GetHistogram().GetSampleCount() > 0 {
				found = true
				break
			}
		}
	}

	assert.True(t, found, "expected http_request_duration_seconds histogram with sample count > 0")
}
