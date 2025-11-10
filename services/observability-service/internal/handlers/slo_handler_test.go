package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"example.com/lma/observability-service/internal/models"
)

type fakeSLOService struct {
	created *models.SLO
	get     *models.SLO
	list    []*models.SLO
	status  *models.SLOStatus
	history []*models.SLOHistory
	err     error
}

func (f *fakeSLOService) CreateSLO(ctx context.Context, slo *models.SLO) error {
	f.created = slo
	return f.err
}
func (f *fakeSLOService) GetSLO(ctx context.Context, id string) (*models.SLO, error) {
	if f.get == nil {
		return nil, f.err
	}
	return f.get, nil
}
func (f *fakeSLOService) ListProjectSLOs(ctx context.Context, projectID string) ([]*models.SLO, error) {
	return f.list, nil
}
func (f *fakeSLOService) UpdateSLO(ctx context.Context, slo *models.SLO) error { return f.err }
func (f *fakeSLOService) DeleteSLO(ctx context.Context, id string) error       { return f.err }
func (f *fakeSLOService) EvaluateSLO(ctx context.Context, sloID string) (*models.SLOStatus, error) {
	return f.status, f.err
}
func (f *fakeSLOService) GetSLOStatus(ctx context.Context, sloID string) (*models.SLOStatus, error) {
	return f.status, f.err
}
func (f *fakeSLOService) calculateCompliance(target float64, actual float64, sloType models.SLOType) float64 {
	return 0
}
func (f *fakeSLOService) calculateBurnRate(current, previous float64, window time.Duration) float64 {
	return 0
}
func (f *fakeSLOService) GetSLOHistory(ctx context.Context, sloID string, from, to time.Time) ([]*models.SLOHistory, error) {
	return f.history, nil
}

func TestSLOHandler_CreateSLO_Success(t *testing.T) {
	svc := &fakeSLOService{}
	h := NewSLOHandler(svc)
	body, _ := json.Marshal(map[string]any{
		"service_name": "api", "type": "availability", "target": 99.9, "window_seconds": 86400, "query": "up"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/proj-1/slos", bytes.NewReader(body))
	h.CreateSLO(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201 got %d", rec.Code)
	}
	if svc.created == nil || svc.created.ProjectID != "proj-1" {
		t.Fatalf("expected created with project id")
	}
}

func TestSLOHandler_CreateSLO_InvalidJSON(t *testing.T) {
	svc := &fakeSLOService{}
	h := NewSLOHandler(svc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/proj-1/slos", bytes.NewReader([]byte("{")))
	h.CreateSLO(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d", rec.Code)
	}
}

func TestSLOHandler_CreateSLO_UnknownField_400(t *testing.T) {
	svc := &fakeSLOService{}
	h := NewSLOHandler(svc)
	body := []byte(`{"service_name":"api","type":"availability","target":99.9,"window_seconds":60,"query":"up","extra":true}`)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/p/slos", bytes.NewReader(body))
	h.CreateSLO(rec, req)
	if rec.Code != http.StatusBadRequest { t.Fatalf("want 400 got %d", rec.Code) }
}

func TestSLOHandler_CreateSLO_TooLarge_413(t *testing.T) {
	svc := &fakeSLOService{}
	h := NewSLOHandler(svc)
big := bytes.Repeat([]byte("a"), 2<<20)
	rec := httptest.NewRecorder()
	body := []byte(`{"service_name":"api","type":"availability","target":99.9,"window_seconds":60,"query":"`+string(big)+`"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/p/slos", bytes.NewReader(body))
	h.CreateSLO(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge { t.Fatalf("want 413 got %d", rec.Code) }
}

func TestSLOHandler_GetSLO_SuccessNotFound(t *testing.T) {
	svc := &fakeSLOService{get: &models.SLO{ID: "s1"}}
	h := NewSLOHandler(svc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/slos/s1", nil)
	h.GetSLO(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", rec.Code)
	}

	svc2 := &fakeSLOService{get: nil, err: context.Canceled}
	h2 := NewSLOHandler(svc2)
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/v1/slos/missing", nil)
	h2.GetSLO(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("want 404 got %d", rec2.Code)
	}
}

func TestSLOHandler_ListProjectSLOs_Empty(t *testing.T) {
	svc := &fakeSLOService{list: []*models.SLO{}}
	h := NewSLOHandler(svc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/projects/proj-1/slos", nil)
	h.ListProjectSLOs(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", rec.Code)
	}
}

func TestSLOHandler_UpdateSLO_SuccessValidation(t *testing.T) {
	svc := &fakeSLOService{err: nil}
	h := NewSLOHandler(svc)
	body, _ := json.Marshal(map[string]any{"service_name": "api", "type": "availability", "target": 99.9, "window_seconds": 86400, "query": "up"})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/v1/slos/s1", bytes.NewReader(body))
	h.UpdateSLO(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", rec.Code)
	}

	svc2 := &fakeSLOService{err: context.DeadlineExceeded}
	h2 := NewSLOHandler(svc2)
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPut, "/v1/slos/s1", bytes.NewReader(body))
	h2.UpdateSLO(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("want 400 got %d", rec2.Code)
	}
}

func TestSLOHandler_DeleteSLO_Success(t *testing.T) {
	svc := &fakeSLOService{}
	h := NewSLOHandler(svc)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v1/slos/s1", nil)
	h.DeleteSLO(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204 got %d", rec.Code)
	}
}

func TestSLOHandler_GetSLOStatus_AndHistory(t *testing.T) {
	svc := &fakeSLOService{status: &models.SLOStatus{SLOID: "s1", Compliance: 99.0}, history: []*models.SLOHistory{{SLOID: "s1", Timestamp: time.Now(), Compliance: 99.0}}}
	h := NewSLOHandler(svc)
	// status
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/slos/s1/status", nil)
	h.GetSLOStatus(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", rec.Code)
	}
	// history
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/v1/slos/s1/history?from=2025-01-01T00:00:00Z&to=2025-01-02T00:00:00Z", nil)
	h.GetSLOHistory(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", rec2.Code)
	}
}
