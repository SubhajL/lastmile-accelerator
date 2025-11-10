import { describe, it, expect } from 'vitest';
import { newDb } from 'pg-mem';
import type { Pool } from 'pg';
import { runMigrations } from '../../../repo/db.js';
import { PgTestRunsRepo } from '../../../repo/test-runs.pg.repo.js';
import { PgBrowserTestRunsRepo } from '../../../repo/browser-test-runs.pg.repo.js';
import { randomUUID } from 'node:crypto';

function createMemPool(): Pool {
  const db = newDb({ autoCreateForeignKeyIndices: true });
  // gen_random_uuid compatibility not needed since we generate app-side
  const adapter = db.adapters.createPg();
  const { Pool: MemPool } = adapter as any;
  return new (MemPool as any)();
}

describe('PgTestRunsRepo & PgBrowserTestRunsRepo', () => {
  it('CRUD, filtering and pagination work', async () => {
    const pool = createMemPool();
    await runMigrations(pool);

    const runs = new PgTestRunsRepo(pool);
    const browsers = new PgBrowserTestRunsRepo(pool);

    const projectId = '11111111-1111-1111-1111-111111111111';
    const snapshotId = '22222222-2222-2222-2222-222222222222';

    const r1 = await runs.create(projectId, { type: 'unit', snapshotId });
    const r2 = await runs.create(projectId, { type: 'e2e' });

    expect(r1.id).toBeDefined();
    expect((await runs.getById(r1.id))?.id).toBe(r1.id);

    // Update status to running then passed with artifacts/results
    await runs.updateStatus(r1.id, 'running', { startedAt: new Date('2024-01-01T00:00:00.000Z') });
    const finished = await runs.updateStatus(r1.id, 'passed', {
      finishedAt: new Date('2024-01-01T00:05:00.000Z'),
      artifacts: { s3: ['artifacts/log.txt'] },
      results: { passed: 10, failed: 0 },
    });
    expect(finished?.status).toBe('passed');

    // List with filtering
    const page1 = await runs.listByProject(projectId, { limit: 1, status: 'queued' });
    expect(page1.items.length).toBe(1);
    const page2 = await runs.listByProject(projectId, { limit: 1, cursor: page1.nextCursor });
    expect(page2.items.length).toBe(1);

    // Browser test runs for r1
    const b1 = await browsers.create(r1.id, { browser: 'chrome', os: 'macOS', viewport: '1280x800' });
    const b2 = await browsers.create(r1.id, { browser: 'firefox' });
    expect((await browsers.getById(b1.id))?.browser).toBe('chrome');

    await browsers.updateStatus(b1.id, 'passed', {
      finishedAt: new Date('2024-01-01T00:03:00.000Z'),
      screenshots: ['s3://bucket/shot1.png'],
    });

    const listB = await browsers.listByTestRun(r1.id);
    expect(listB.items.length).toBe(2);
    expect(listB.items.some(i => i.status === 'passed')).toBe(true);
  });
});
