import { describe, it, expect } from 'vitest';
import { newDb } from 'pg-mem';
import type { Pool } from 'pg';
import { runMigrations } from '../../../repo/db.js';
import { PgPreviewEnvsRepo } from '../../../repo/preview-envs.pg.repo.js';

function createMemPool(): Pool {
  const db = newDb({ autoCreateForeignKeyIndices: true });
  const adapter = db.adapters.createPg();
  const { Pool: MemPool } = adapter as any;
  return new (MemPool as any)();
}

describe('PgPreviewEnvsRepo', () => {
  it('CRUD and pagination work', async () => {
    const pool = createMemPool();
    await runMigrations(pool);

    const repo = new PgPreviewEnvsRepo(pool);
    const projectId = '11111111-1111-1111-1111-111111111111';
    const snapshotId = '22222222-2222-2222-2222-222222222222';

    const p1 = await repo.create(projectId, { url: 'https://p1.example', snapshotId });
    const p2 = await repo.create(projectId, { url: 'https://p2.example' });

    expect(p1.id).toBeDefined();
    expect((await repo.getById(p1.id))?.url).toBe('https://p1.example');

    // update status and expires
    const updated = await repo.update(p1.id, { status: 'ready', expiresAt: new Date('2024-01-02T00:00:00.000Z') });
    expect(updated?.status).toBe('ready');

    // list with pagination
    const page1 = await repo.listByProject(projectId, { limit: 1 });
    expect(page1.items.length).toBe(1);
    const page2 = await repo.listByProject(projectId, { limit: 1, cursor: page1.nextCursor });
    expect(page2.items.length).toBe(1);

    // delete
    const ok = await repo.delete(p2.id);
    expect(ok).toBe(true);
    expect(await repo.getById(p2.id)).toBeNull();
  });
});
