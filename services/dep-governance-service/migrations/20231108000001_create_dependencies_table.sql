-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create dependencies table
CREATE TABLE IF NOT EXISTS dependencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id UUID NOT NULL,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    ecosystem TEXT NOT NULL,
    is_direct BOOLEAN NOT NULL DEFAULT false,
    parent_id UUID REFERENCES dependencies(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Ensure no duplicate dependencies per snapshot
    CONSTRAINT unique_dependency_per_snapshot UNIQUE (snapshot_id, name, version, ecosystem)
);

-- Index for querying dependencies by snapshot
CREATE INDEX IF NOT EXISTS idx_dependencies_snapshot_id ON dependencies(snapshot_id);

-- Index for direct vs transitive queries
CREATE INDEX IF NOT EXISTS idx_dependencies_snapshot_direct ON dependencies(snapshot_id, is_direct);

-- Index for name lookups (GIN for pattern matching)
CREATE INDEX IF NOT EXISTS idx_dependencies_name ON dependencies USING GIN (name gin_trgm_ops);

-- Index for tree traversal
CREATE INDEX IF NOT EXISTS idx_dependencies_parent_id ON dependencies(parent_id) WHERE parent_id IS NOT NULL;
