import { describe, it, expect } from 'vitest';
import { newDb } from 'pg-mem';
import type { Pool } from 'pg';
import { runMigrations } from '../../../repo/db.js';

function createMemPool(): Pool {
  const db = newDb({ autoCreateForeignKeyIndices: true });
  const adapter = db.adapters.createPg();
  const { Pool: MemPool } = adapter;
  return new (MemPool as any)();
}

describe('repo/migrations', () => {
  it('applies base schema and is idempotent', async () => {
    const pool = createMemPool();

    await runMigrations(pool);
    // Running again should not throw
    await runMigrations(pool);

    // Try to use tables to ensure they exist
    await pool.query(`INSERT INTO test_scaffolds (id, project_id, type, framework, language, config) VALUES ('00000000-0000-0000-0000-000000000001','11111111-1111-1111-1111-111111111111','unit','vitest','ts','{}')`);
    const res = await pool.query(`SELECT * FROM test_scaffolds`);
    expect(res.rowCount).toBe(1);
  });
});
