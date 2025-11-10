CREATE TABLE IF NOT EXISTS secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL,
    key VARCHAR(255) NOT NULL,
    environment VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    UNIQUE(project_id, key, environment)
);

CREATE INDEX IF NOT EXISTS idx_secrets_project_env ON secrets(project_id, environment);
CREATE INDEX IF NOT EXISTS idx_secrets_tenant ON secrets(tenant_id);

CREATE TABLE IF NOT EXISTS secret_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    secret_id UUID NOT NULL REFERENCES secrets(id) ON DELETE CASCADE,
    version_number INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by VARCHAR(255) NOT NULL,
    rotated_at TIMESTAMPTZ,
    UNIQUE(secret_id, version_number)
);

CREATE INDEX IF NOT EXISTS idx_secret_versions_secret ON secret_versions(secret_id);

CREATE TABLE IF NOT EXISTS env_parity_checks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id VARCHAR(255) NOT NULL,
    scan_timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    missing_keys JSONB,
    mismatched_keys JSONB,
    extra_keys JSONB,
    has_drift BOOLEAN NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_env_parity_project ON env_parity_checks(project_id, scan_timestamp DESC);

CREATE TABLE IF NOT EXISTS client_leak_scans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id VARCHAR(255) NOT NULL,
    snapshot_id VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    line_number INT NOT NULL,
    pattern VARCHAR(255) NOT NULL,
    severity VARCHAR(50) NOT NULL,
    fixed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_client_leak_scans_project ON client_leak_scans(project_id);
CREATE INDEX IF NOT EXISTS idx_client_leak_scans_snapshot ON client_leak_scans(snapshot_id);

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id VARCHAR(255) NOT NULL,
    project_id VARCHAR(255) NOT NULL,
    secret_key VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL,
    actor VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_project ON audit_logs(project_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant ON audit_logs(tenant_id, timestamp DESC);
