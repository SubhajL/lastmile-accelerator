import { describe, it, expect } from 'vitest';
import { newDb } from 'pg-mem';
import type { Pool } from 'pg';
import { runMigrations } from '../../../repo/db.js';
import { PgScaffoldsRepo } from '../../../repo/scaffolds.pg.repo.js';
import { randomUUID } from 'node:crypto';

function createMemPool(): Pool {
  const db = newDb({ autoCreateForeignKeyIndices: true });
  // emulate pgcrypto gen_random_uuid
// @ts-ignore pg-mem type quirk for returns type
db.public.registerFunction({ name: 'gen_random_uuid', implementation: () => randomUUID() });
  const adapter = db.adapters.createPg();
  const { Pool: MemPool } = adapter;
  return new (MemPool as any)();
}

describe('PgScaffoldsRepo', () => {
  it('CRUD and pagination work', async () => {
    const pool = createMemPool();
    await runMigrations(pool);
    const repo = new PgScaffoldsRepo(pool);

    const projectId = '11111111-1111-1111-1111-111111111111';
    const a = await repo.create(projectId, { type: 'unit', framework: 'vitest', language: 'ts', config: { a: 1 } });
    const b = await repo.create(projectId, { type: 'unit', framework: 'vitest', language: 'ts', config: { b: 2 } });

    const gotA = await repo.getById(a.id);
    expect(gotA?.id).toBe(a.id);

    const page1 = await repo.listByProject(projectId, 1);
    expect(page1.items.length).toBe(1);
    expect(page1.nextCursor).toBeDefined();

    const page2 = await repo.listByProject(projectId, 1, page1.nextCursor);
    expect(page2.items.length).toBe(1);

    const upd = await repo.update(a.id, { framework: 'jest' });
    expect(upd?.framework).toBe('jest');

    const del = await repo.delete(a.id);
    expect(del).toBe(true);
    const missing = await repo.getById(a.id);
    expect(missing).toBeNull();
  });
});
