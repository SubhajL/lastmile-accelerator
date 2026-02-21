import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { join } from 'path';

describe('db/outbox schema DDL', () => {
  it('contains outbox table with PK, status check, and index', () => {
    const p = join(process.cwd(), 'src', 'db', 'schema.sql');
    const sql = readFileSync(p, 'utf8');

    expect(sql).toContain('CREATE TABLE IF NOT EXISTS notification_outbox_messages');
    expect(sql).toContain('(\n    dedup_key TEXT PRIMARY KEY');
    expect(sql).toContain("status TEXT NOT NULL CHECK (status IN ('pending','sent','failed'))");
    expect(sql).toContain('created_at TIMESTAMPTZ NOT NULL DEFAULT now()');
    expect(sql).toContain('CREATE INDEX IF NOT EXISTS idx_outbox_created_at ON notification_outbox_messages(created_at)');
  });
});
