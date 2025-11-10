package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/storage"
)

func TestOTelRepository_ListPresets_ReturnsFive(t *testing.T) {
	repo := NewOTelRepository(nil)
	presets, err := repo.ListPresets(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(presets) != 5 {
		t.Fatalf("expected 5 presets, got %d", len(presets))
	}
}

func TestOTelRepository_GetPreset_KnownFramework(t *testing.T) {
	repo := NewOTelRepository(nil)
	preset, err := repo.GetPreset(context.Background(), models.FrameworkGo)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if preset.Framework != models.FrameworkGo {
		t.Errorf("expected framework go, got %s", preset.Framework)
	}
}

func TestOTelRepository_GetPreset_UnknownFramework(t *testing.T) {
	repo := NewOTelRepository(nil)
	_, err := repo.GetPreset(context.Background(), models.Framework("unknown"))
	if err == nil {
		t.Fatal("expected error for unknown framework")
	}
}

func TestOTelRepository_SaveAndGetProjectConfig(t *testing.T) {
	mdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mdb.Close()

	pdb := storage.NewPostgresDBFromSQL(mdb)
	repo := NewOTelRepository(pdb)

	cfg := &models.ProjectOTelConfig{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		ProjectID: "proj-123",
		Framework: models.FrameworkNextJS,
		Config: map[string]interface{}{
			"trace_endpoint": "http://collector:4318",
			"sampling_rate":  0.2,
		},
	}

	// Expect upsert
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO project_otel_configs (id, project_id, framework, config, applied_at, updated_at) VALUES ($1,$2,$3,$4,NOW(),NOW()) ON CONFLICT (project_id) DO UPDATE SET framework = EXCLUDED.framework, config = EXCLUDED.config, updated_at = NOW()")).
		WithArgs(cfg.ID, cfg.ProjectID, string(cfg.Framework), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := repo.SaveProjectConfig(context.Background(), cfg); err != nil {
		t.Fatalf("SaveProjectConfig error: %v", err)
	}

	// Expect select
	bytes, _ := json.Marshal(cfg.Config)
	rows := sqlmock.NewRows([]string{"id", "project_id", "framework", "config", "applied_at", "updated_at"}).
		AddRow(cfg.ID, cfg.ProjectID, string(cfg.Framework), string(bytes), sql.NullTime{}, sql.NullTime{})
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id, framework, config, applied_at, updated_at FROM project_otel_configs WHERE project_id = $1")).
		WithArgs(cfg.ProjectID).
		WillReturnRows(rows)

	got, err := repo.GetProjectConfig(context.Background(), cfg.ProjectID)
	if err != nil {
		t.Fatalf("GetProjectConfig error: %v", err)
	}
	if got.ProjectID != cfg.ProjectID {
		t.Errorf("expected project_id %s, got %s", cfg.ProjectID, got.ProjectID)
	}
}

func TestOTelRepository_GetProjectConfig_NotFound(t *testing.T) {
	mdb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer mdb.Close()
	pdb := storage.NewPostgresDBFromSQL(mdb)
	repo := NewOTelRepository(pdb)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, project_id, framework, config, applied_at, updated_at FROM project_otel_configs WHERE project_id = $1")).
		WithArgs("missing").
		WillReturnError(errors.New("no rows in result set"))

	_, err = repo.GetProjectConfig(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected not found error")
	}
}
