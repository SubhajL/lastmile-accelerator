package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetupRoutes_HealthzOK(t *testing.T) {
	r := SetupRoutes(nil, nil, nil)
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil { t.Fatalf("request failed: %v", err) }
	if resp.StatusCode != 200 { t.Fatalf("expected 200, got %d", resp.StatusCode) }
}
