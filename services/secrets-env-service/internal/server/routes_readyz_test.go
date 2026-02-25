package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/lma/secrets-env-service/internal/handlers"
	"example.com/lma/secrets-env-service/internal/startup"
	"github.com/stretchr/testify/assert"
)

func TestSetupRoutes_Readyz_Returns200WhenReady(t *testing.T) {
	readiness := startup.NewReadiness(map[string]startup.CheckFunc{
		"vault": func(ctx context.Context) error { return nil },
	}, nil)

	r := SetupRoutes(nil, nil, nil, handlers.ReadyCheck(readiness))
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/readyz")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSetupRoutes_Readyz_Returns503WhenNotReady(t *testing.T) {
	readiness := startup.NewReadiness(map[string]startup.CheckFunc{
		"vault": func(ctx context.Context) error { return assert.AnError },
	}, nil)

	r := SetupRoutes(nil, nil, nil, handlers.ReadyCheck(readiness))
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/readyz")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, false, body["ready"])
}
