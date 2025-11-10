CREATE TABLE IF NOT EXISTS slos (
    id UUID PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    service_name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    target DECIMAL(5,2) NOT NULL CHECK (target >= 0 AND target <= 100),
    window_seconds INT NOT NULL CHECK (window_seconds > 0),
    query TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_slos_project_id ON slos(project_id);

CREATE TABLE IF NOT EXISTS slo_status (
    slo_id UUID PRIMARY KEY REFERENCES slos(id) ON DELETE CASCADE,
    compliance DECIMAL(5,2) NOT NULL,
    burn_rate DECIMAL(10,4) NOT NULL,
    remaining_budget DECIMAL(5,2) NOT NULL,
    last_calculated TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS slo_history (
    id UUID PRIMARY KEY,
    slo_id UUID NOT NULL REFERENCES slos(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    compliance DECIMAL(5,2) NOT NULL,
    burn_rate DECIMAL(10,4) NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_slo_history_slo_timestamp ON slo_history(slo_id, timestamp DESC);
