package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/lma/secrets-env-service/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeParityService struct {
	latest *domain.EnvParityCheck
	history []*domain.EnvParityCheck
}

func (f *fakeParityService) CheckParity(_ any, projectID, baseEnv, compareEnv string) (*domain.EnvParityCheck, error) {
	return &domain.EnvParityCheck{ProjectID: projectID, MissingKeys: []string{"A"}, ExtraKeys: []string{"C"}}, nil
}
func (f *fakeParityService) GetLatestCheck(_ any, projectID string) (*domain.EnvParityCheck, error) {
	if f.latest != nil { return f.latest, nil }
	return nil, assert.AnError
}
func (f *fakeParityService) GetCheckHistory(_ any, projectID string, limit int) ([]*domain.EnvParityCheck, error) {
	return f.history, nil
}

func TestParityHandler_CheckParity_Valid(t *testing.T) {
	svc := &fakeParityService{}
	h := NewParityHandler(svc)
	r := chi.NewRouter()
	r.Post("/v1/projects/{projectID}/env-parity", h.CheckParity)

	body := map[string]any{"baseEnv":"dev","compareEnv":"prod"}
	b,_ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/proj-1/env-parity", bytes.NewReader(b))
	req.Header.Set("Content-Type","application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestParityHandler_GetLatest_NotFound(t *testing.T) {
	svc := &fakeParityService{}
	h := NewParityHandler(svc)
	r := chi.NewRouter()
	r.Get("/v1/projects/{projectID}/env-parity/latest", h.GetLatestCheck)

	req := httptest.NewRequest(http.MethodGet, "/v1/projects/p/env-parity/latest", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestParityHandler_GetHistory_LimitApplied(t *testing.T) {
	svc := &fakeParityService{history: []*domain.EnvParityCheck{{ProjectID:"p"},{ProjectID:"p"}}}
	h := NewParityHandler(svc)
	r := chi.NewRouter()
	r.Get("/v1/projects/{projectID}/env-parity/history", h.GetCheckHistory)

	req := httptest.NewRequest(http.MethodGet, "/v1/projects/p/env-parity/history?limit=1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
}
