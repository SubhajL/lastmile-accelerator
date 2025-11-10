CREATE TABLE IF NOT EXISTS project_otel_configs (
    id UUID PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL UNIQUE,
    framework VARCHAR(50) NOT NULL,
    config JSONB NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_project_otel_project_id ON project_otel_configs(project_id);
