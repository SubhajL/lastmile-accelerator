CREATE TABLE IF NOT EXISTS error_groups (
    id UUID PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL,
    fingerprint VARCHAR(64) NOT NULL,
    title TEXT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'open',
    first_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    occurrences BIGINT NOT NULL DEFAULT 1,
    sample_stack TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_error_groups_project_fingerprint ON error_groups(project_id, fingerprint);
CREATE INDEX IF NOT EXISTS idx_error_groups_project_lastseen ON error_groups(project_id, last_seen DESC);

CREATE TABLE IF NOT EXISTS error_events (
    id UUID PRIMARY KEY,
    group_id UUID NOT NULL REFERENCES error_groups(id) ON DELETE CASCADE,
    project_id VARCHAR(255) NOT NULL,
    ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    message TEXT NOT NULL,
    stack TEXT,
    metadata JSONB
);
CREATE INDEX IF NOT EXISTS idx_error_events_group_ts ON error_events(group_id, ts DESC);
