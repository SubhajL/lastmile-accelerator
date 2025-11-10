package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/lma/observability-service/internal/models"
)

type fakeOTelService struct {
	presets []*models.OTelPreset
	applied *models.ProjectOTelConfig
}

func (f *fakeOTelService) GetAvailablePresets(ctx context.Context) ([]*models.OTelPreset, error) {
	return f.presets, nil
}
func (f *fakeOTelService) GetPresetForFramework(ctx context.Context, framework models.Framework) (*models.OTelPreset, error) {
	for _, p := range f.presets {
		if p.Framework == framework {
			return p, nil
		}
	}
	return nil, fmt.Errorf("not found")
}
func (f *fakeOTelService) ApplyPresetToProject(ctx context.Context, projectID string, framework models.Framework) (*models.ProjectOTelConfig, error) {
	f.applied = &models.ProjectOTelConfig{ProjectID: projectID, Framework: framework}
	return f.applied, nil
}
func (f *fakeOTelService) GetProjectConfiguration(ctx context.Context, projectID string) (*models.ProjectOTelConfig, error) {
	return f.applied, nil
}

func TestOTelHandler_GetPresets(t *testing.T) {
	svc := &fakeOTelService{presets: []*models.OTelPreset{{Framework: models.FrameworkGo}}}
	h := NewOTelHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/v1/otel/presets", nil)
	rec := httptest.NewRecorder()
	h.GetPresets(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", rec.Code)
	}
}

func TestOTelHandler_ApplyAndGetProjectConfig(t *testing.T) {
	svc := &fakeOTelService{presets: []*models.OTelPreset{{Framework: models.FrameworkGo}}}
	h := NewOTelHandler(svc)
	body, _ := json.Marshal(map[string]string{"framework": "go"})
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/proj-1/otel/config", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ApplyConfigToProject(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201 got %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/v1/projects/proj-1/otel/config", nil)
	rec2 := httptest.NewRecorder()
	h.GetProjectConfig(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("want 200 got %d", rec2.Code)
	}
}

func TestOTelHandler_Apply_UnknownField_400(t *testing.T) {
	svc := &fakeOTelService{}
	h := NewOTelHandler(svc)
	rec := httptest.NewRecorder()
	body := []byte(`{"framework":"go","extra":true}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/projects/proj-1/otel/config", bytes.NewReader(body))
	h.ApplyConfigToProject(rec, req)
	if rec.Code != http.StatusBadRequest { t.Fatalf("want 400 got %d", rec.Code) }
}
