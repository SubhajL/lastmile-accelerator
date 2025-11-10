-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS db_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    name TEXT NOT NULL,
    driver TEXT NOT NULL CHECK (driver IN ('postgres', 'mysql', 'sqlite')),
    dsn_ref TEXT NOT NULL, -- Vault path or reference, NOT the actual DSN
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_default_per_project UNIQUE (project_id, is_default) 
        WHERE is_default = TRUE,
    CONSTRAINT unique_name_per_project UNIQUE (project_id, name)
);

CREATE INDEX idx_db_connections_project_id ON db_connections(project_id);

CREATE TABLE IF NOT EXISTS role_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL UNIQUE,
    spec_yaml TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS migration_audits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    migration_name TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pass', 'warn', 'fail')),
    findings_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_migration_audits_project_created ON migration_audits(project_id, created_at DESC);

CREATE TABLE IF NOT EXISTS index_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL,
    table_name TEXT NOT NULL,
    column_names TEXT[] NOT NULL,
    reason TEXT,
    benefit_score INTEGER CHECK (benefit_score >= 0 AND benefit_score <= 100),
    applied BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_recommendation UNIQUE (project_id, table_name, column_names)
);

CREATE INDEX idx_index_recommendations_project ON index_recommendations(project_id);
CREATE INDEX idx_index_recommendations_unapplied ON index_recommendations(project_id, applied) 
    WHERE applied = FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS index_recommendations;
DROP TABLE IF EXISTS migration_audits;
DROP TABLE IF EXISTS role_policies;
DROP TABLE IF EXISTS db_connections;
-- +goose StatementEnd
