import type { Pool } from 'pg';
import { MIGRATIONS } from './migrations/index.js';

export async function runMigrations(pool: Pool): Promise<void> {
  try {
    await pool.query(`SELECT 1 FROM schema_migrations LIMIT 1`);
  } catch {
    await pool.query(`CREATE TABLE schema_migrations (id text PRIMARY KEY, applied_at timestamp NOT NULL DEFAULT now())`);
  }
  const res = await pool.query<{ id: string }>(`SELECT id FROM schema_migrations`);
  const applied = new Set(res.rows.map((r) => r.id));

  for (const m of MIGRATIONS) {
    if (applied.has(m.id)) continue;
    await pool.query('BEGIN');
    try {
      await pool.query(m.sql);
      await pool.query(`INSERT INTO schema_migrations (id) VALUES ($1)`, [m.id]);
      await pool.query('COMMIT');
    } catch (e) {
      await pool.query('ROLLBACK');
      throw e;
    }
  }
}
