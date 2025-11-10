-- 001_init.sql: initial schema for projects-service

-- Create tenants table
CREATE TABLE IF NOT EXISTS tenants (
  id UUID PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  slug VARCHAR(255) UNIQUE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);

-- Create projects table
CREATE TABLE IF NOT EXISTS projects (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW(),
  deleted_at TIMESTAMPTZ NULL
);

-- Create members table
CREATE TABLE IF NOT EXISTS members (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id),
  user_id UUID NOT NULL,
  email VARCHAR(255) NOT NULL,
  role VARCHAR(50) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create environments table
CREATE TABLE IF NOT EXISTS environments (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL REFERENCES projects(id),
  name VARCHAR(50) NOT NULL,
  config JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create ingestion_modes table
CREATE TABLE IF NOT EXISTS ingestion_modes (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL REFERENCES projects(id),
  mode VARCHAR(10) NOT NULL,
  is_default BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for tenant isolation
CREATE INDEX IF NOT EXISTS idx_projects_tenant_id ON projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_members_tenant_id ON members(tenant_id);
CREATE INDEX IF NOT EXISTS idx_environments_project_id ON environments(project_id);
CREATE INDEX IF NOT EXISTS idx_ingestion_modes_project_id ON ingestion_modes(project_id);
