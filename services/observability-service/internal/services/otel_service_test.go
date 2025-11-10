package services

import (
	"context"
	"errors"
	"testing"

	"example.com/lma/observability-service/internal/models"
	"go.opentelemetry.io/otel/trace"
)

type fakeOTelRepo struct {
	presets []*models.OTelPreset
	saved   *models.ProjectOTelConfig
	cfg     *models.ProjectOTelConfig
	getErr  error
}

func (f *fakeOTelRepo) GetPreset(ctx context.Context, fw models.Framework) (*models.OTelPreset, error) {
	for _, p := range f.presets {
		if p.Framework == fw {
			return p, nil
		}
	}
	return nil, errors.New("not found")
}
func (f *fakeOTelRepo) ListPresets(ctx context.Context) ([]*models.OTelPreset, error) {
	return f.presets, nil
}
func (f *fakeOTelRepo) SaveProjectConfig(ctx context.Context, cfg *models.ProjectOTelConfig) error {
	f.saved = cfg
	return nil
}
func (f *fakeOTelRepo) GetProjectConfig(ctx context.Context, pid string) (*models.ProjectOTelConfig, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.cfg, nil
}

func TestOTelService_BasicFlows(t *testing.T) {
	repo := &fakeOTelRepo{presets: []*models.OTelPreset{{Framework: models.FrameworkGo}}}
	var noop trace.Tracer
	svc := NewOTelService(repo, noop)

	// List presets
	ps, err := svc.GetAvailablePresets(context.Background())
	if err != nil || len(ps) != 1 {
		t.Fatalf("expected 1 preset, got %v, err=%v", len(ps), err)
	}

	// Get preset
	p, err := svc.GetPresetForFramework(context.Background(), models.FrameworkGo)
	if err != nil || p.Framework != models.FrameworkGo {
		t.Fatalf("get preset failed: %v", err)
	}

	// Apply preset
	applied, err := svc.ApplyPresetToProject(context.Background(), "proj-1", models.FrameworkGo)
	if err != nil || applied.ProjectID != "proj-1" {
		t.Fatalf("apply failed: %v", err)
	}
	if repo.saved == nil {
		t.Fatalf("expected save to be called")
	}

	// Get project config
	repo.cfg = applied
	got, err := svc.GetProjectConfiguration(context.Background(), "proj-1")
	if err != nil || got.ProjectID != "proj-1" {
		t.Fatalf("get project cfg failed: %v", err)
	}
}
