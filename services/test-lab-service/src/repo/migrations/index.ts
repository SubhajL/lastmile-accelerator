export const MIGRATIONS: { id: string; sql: string }[] = [
  {
    id: '0001_init',
    sql: `
CREATE TABLE IF NOT EXISTS test_scaffolds (
  id uuid PRIMARY KEY,
  project_id uuid NOT NULL,
  type text NOT NULL,
  framework text NOT NULL,
  language text NOT NULL,
  config jsonb NOT NULL DEFAULT '{}',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_test_scaffolds_project_created ON test_scaffolds (project_id, created_at DESC);

CREATE TABLE IF NOT EXISTS test_runs (
  id uuid PRIMARY KEY,
  project_id uuid NOT NULL,
  snapshot_id uuid,
  type text NOT NULL,
  status text NOT NULL,
  started_at timestamptz,
  finished_at timestamptz,
  results jsonb,
  artifacts jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_test_runs_project_created ON test_runs (project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_test_runs_snapshot_status ON test_runs (snapshot_id, status);

CREATE TABLE IF NOT EXISTS browser_test_runs (
  id uuid PRIMARY KEY,
  test_run_id uuid NOT NULL REFERENCES test_runs(id) ON DELETE CASCADE,
  browser text NOT NULL,
  os text,
  viewport text,
  status text NOT NULL,
  screenshots jsonb,
  logs jsonb,
  started_at timestamptz,
  finished_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_browser_test_runs_run ON browser_test_runs (test_run_id);

CREATE TABLE IF NOT EXISTS preview_environments (
  id uuid PRIMARY KEY,
  project_id uuid NOT NULL,
  snapshot_id uuid,
  url text NOT NULL,
  status text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_preview_envs_project_created ON preview_environments (project_id, created_at DESC);
`,
  },
];
