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
)

type fakeLeakService struct { scans []*domain.ClientLeakScan }

func (f *fakeLeakService) ScanSnapshot(_ any, projectID, snapshotID string) ([]*domain.ClientLeakScan, error) {
	f.scans = []*domain.ClientLeakScan{{ID:"1"},{ID:"2"}}
	return f.scans, nil
}
func (f *fakeLeakService) GetScanResults(_ any, projectID, snapshotID, severity string) ([]*domain.ClientLeakScan, error) { return f.scans, nil }
func (f *fakeLeakService) MarkAsFixed(_ any, scanID string) error { return nil }

func TestLeakScanHandler_ScanSnapshot_Valid(t *testing.T) {
	svc := &fakeLeakService{}
	h := NewLeakScanHandler(svc)
	r := chi.NewRouter()
	r.Post("/v1/projects/{projectID}/scan/client-leaks", h.ScanSnapshot)
	b,_ := json.Marshal(map[string]any{"snapshotID":"s1"})
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/p/scan/client-leaks", bytes.NewReader(b))
	req.Header.Set("Content-Type","application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestLeakScanHandler_GetScanResults(t *testing.T) {
	svc := &fakeLeakService{scans: []*domain.ClientLeakScan{{ID:"1"}}}
	h := NewLeakScanHandler(svc)
	r := chi.NewRouter()
	r.Get("/v1/projects/{projectID}/scan/client-leaks/{snapshotID}", h.GetScanResults)
	req := httptest.NewRequest(http.MethodGet, "/v1/projects/p/scan/client-leaks/s1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestLeakScanHandler_MarkAsFixed(t *testing.T) {
	svc := &fakeLeakService{}
	h := NewLeakScanHandler(svc)
	r := chi.NewRouter()
	r.Patch("/v1/projects/{projectID}/scan/client-leaks/{scanID}/fix", h.MarkAsFixed)
	req := httptest.NewRequest(http.MethodPatch, "/v1/projects/p/scan/client-leaks/x/fix", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}
