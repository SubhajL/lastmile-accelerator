-- Enable extension for UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Create CVEs table
CREATE TABLE IF NOT EXISTS cves (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cve_id VARCHAR(32) NOT NULL UNIQUE,
    severity TEXT NOT NULL,
    description TEXT,
    published_at TIMESTAMPTZ,
    source TEXT,
    cvss_score NUMERIC(3,1),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create dependency_vulnerabilities junction table
CREATE TABLE IF NOT EXISTS dependency_vulnerabilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dependency_id UUID NOT NULL,
    cve_id UUID NOT NULL,
    affected_version_range TEXT,
    fixed_version TEXT,
    discovered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'open',
    CONSTRAINT uq_dep_cve UNIQUE (dependency_id, cve_id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_cves_severity ON cves(severity);
CREATE INDEX IF NOT EXISTS idx_dep_vuln_dependency ON dependency_vulnerabilities(dependency_id);
CREATE INDEX IF NOT EXISTS idx_dep_vuln_status ON dependency_vulnerabilities(status);
