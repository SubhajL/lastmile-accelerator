package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck_ReturnsOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestHealthCheck_AcceptsAnyMethod(t *testing.T) {
	methods := []string{http.MethodGet, http.MethodHead, http.MethodPost}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/healthz", nil)
		w := httptest.NewRecorder()

		HealthCheck(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}
