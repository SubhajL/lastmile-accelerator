-- Enable extension for UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Create sboms table
CREATE TABLE IF NOT EXISTS sboms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    snapshot_id UUID NOT NULL,
    format TEXT NOT NULL,
    storage_key TEXT NOT NULL,
    file_hash VARCHAR(64) NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_sbom_storage_key UNIQUE (storage_key)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_sboms_snapshot_id ON sboms(snapshot_id);
CREATE INDEX IF NOT EXISTS idx_sboms_generated_at ON sboms(generated_at DESC);
