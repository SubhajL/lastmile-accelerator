package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"example.com/lma/observability-service/internal/models"
	"example.com/lma/observability-service/internal/storage"
)

type OTelRepository interface {
	GetPreset(ctx context.Context, framework models.Framework) (*models.OTelPreset, error)
	ListPresets(ctx context.Context) ([]*models.OTelPreset, error)
	SaveProjectConfig(ctx context.Context, config *models.ProjectOTelConfig) error
	GetProjectConfig(ctx context.Context, projectID string) (*models.ProjectOTelConfig, error)
}

type postgresOTelRepo struct {
	db *storage.PostgresDB
}

func NewOTelRepository(db *storage.PostgresDB) OTelRepository {
	return &postgresOTelRepo{db: db}
}

var presetCatalog = map[models.Framework]*models.OTelPreset{
	models.FrameworkNextJS: {
		Framework:      models.FrameworkNextJS,
		TraceEndpoint:  "http://otel-collector:4318/v1/traces",
		MetricEndpoint: "http://otel-collector:4318/v1/metrics",
		SamplingRate:   0.1,
		ExporterType:   "otlp",
	},
	models.FrameworkGo: {
		Framework:      models.FrameworkGo,
		TraceEndpoint:  "http://otel-collector:4318/v1/traces",
		MetricEndpoint: "http://otel-collector:4318/v1/metrics",
		SamplingRate:   1.0,
		ExporterType:   "otlp",
	},
	models.FrameworkPython: {
		Framework:      models.FrameworkPython,
		TraceEndpoint:  "http://otel-collector:4318/v1/traces",
		MetricEndpoint: "http://otel-collector:4318/v1/metrics",
		SamplingRate:   0.5,
		ExporterType:   "otlp",
	},
	models.FrameworkNodeJS: {
		Framework:      models.FrameworkNodeJS,
		TraceEndpoint:  "http://otel-collector:4318/v1/traces",
		MetricEndpoint: "http://otel-collector:4318/v1/metrics",
		SamplingRate:   0.2,
		ExporterType:   "otlp",
	},
	models.FrameworkRust: {
		Framework:      models.FrameworkRust,
		TraceEndpoint:  "http://otel-collector:4318/v1/traces",
		MetricEndpoint: "http://otel-collector:4318/v1/metrics",
		SamplingRate:   0.2,
		ExporterType:   "otlp",
	},
}

func (r *postgresOTelRepo) GetPreset(ctx context.Context, framework models.Framework) (*models.OTelPreset, error) {
	preset, ok := presetCatalog[framework]
	if !ok {
		return nil, fmt.Errorf("unknown framework: %s", framework)
	}
	// return a copy to avoid external mutation
	copy := *preset
	return &copy, nil
}

func (r *postgresOTelRepo) ListPresets(ctx context.Context) ([]*models.OTelPreset, error) {
	out := make([]*models.OTelPreset, 0, len(presetCatalog))
	for _, p := range presetCatalog {
		copy := *p
		out = append(out, &copy)
	}
	return out, nil
}

func (r *postgresOTelRepo) SaveProjectConfig(ctx context.Context, config *models.ProjectOTelConfig) error {
	if r.db == nil {
		return fmt.Errorf("db not initialized")
	}
	b, err := json.Marshal(config.Config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	const q = `INSERT INTO project_otel_configs (id, project_id, framework, config, applied_at, updated_at)
	VALUES ($1,$2,$3,$4,NOW(),NOW())
	ON CONFLICT (project_id) DO UPDATE SET framework = EXCLUDED.framework, config = EXCLUDED.config, updated_at = NOW()`
	_, err = r.db.DB().ExecContext(ctx, q, config.ID, config.ProjectID, string(config.Framework), string(b))
	if err != nil {
		return fmt.Errorf("upsert project otel config: %w", err)
	}
	return nil
}

func (r *postgresOTelRepo) GetProjectConfig(ctx context.Context, projectID string) (*models.ProjectOTelConfig, error) {
	if r.db == nil {
		return nil, fmt.Errorf("db not initialized")
	}
	const q = `SELECT id, project_id, framework, config, applied_at, updated_at FROM project_otel_configs WHERE project_id = $1`
	row := r.db.DB().QueryRowContext(ctx, q, projectID)
	var (
		id, pid, fw, cfgStr  string
		appliedAt, updatedAt sql.NullTime
	)
	if err := row.Scan(&id, &pid, &fw, &cfgStr, &appliedAt, &updatedAt); err != nil {
		return nil, fmt.Errorf("get project otel config: %w", err)
	}
	cfg := map[string]interface{}{}
	if err := json.Unmarshal([]byte(cfgStr), &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	return &models.ProjectOTelConfig{
		ID:        id,
		ProjectID: pid,
		Framework: models.Framework(fw),
		Config:    cfg,
		AppliedAt: zeroOr(appliedAt.Time),
		UpdatedAt: zeroOr(updatedAt.Time),
	}, nil
}

func zeroOr(t time.Time) time.Time { return t }
