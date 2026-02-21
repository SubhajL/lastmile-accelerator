import { OUTBOX_TABLE } from './types.js';
import { withSpan } from '../telemetry/tracing.js';

type DbLike = { query: (text: string, params?: unknown[]) => Promise<{ rows: unknown[] }> };

export async function upsertOutboxPending(db: DbLike, dedupKey: string): Promise<boolean> {
  return withSpan('outbox.pg.upsertPending', { key: dedupKey }, async () => {
    const sql = `INSERT INTO ${OUTBOX_TABLE} (dedup_key, status)
                 VALUES ($1, 'pending')
                 ON CONFLICT (dedup_key) DO NOTHING
                 RETURNING 1`;
    const res = await db.query(sql, [dedupKey]);
    return Array.isArray(res.rows) && res.rows.length > 0;
  });
}
