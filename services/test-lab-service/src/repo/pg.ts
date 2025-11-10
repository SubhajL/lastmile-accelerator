import type { Pool } from 'pg';
import { Pool as PgPool } from 'pg';
import { newDb } from 'pg-mem';

export type AnyPool = Pool;

export async function connectDb(connectionString: string): Promise<AnyPool> {
  if (connectionString.startsWith('pgmem://')) {
    const db = newDb({ autoCreateForeignKeyIndices: true });
    // emulate gen_random_uuid
    const { randomUUID } = await import('node:crypto');
// @ts-ignore pg-mem type quirk for returns type
db.public.registerFunction({ name: 'gen_random_uuid', implementation: () => randomUUID() });
    const adapter = db.adapters.createPg();
    const { Pool: MemPool } = adapter as any;
    const pool = new (MemPool as any)();
    return pool as unknown as Pool;
  }
  return new PgPool({ connectionString });
}

export async function closeDb(pool: AnyPool): Promise<void> {
  await (pool as any).end?.();
}
