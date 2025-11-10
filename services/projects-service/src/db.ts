import { Pool, type PoolClient, type QueryResult } from 'pg';

export interface DbConfig {
  connectionString: string;
  max?: number;
  idleTimeoutMillis?: number;
}

let pool: Pool | null = null;

// Factory indirection to make unit testing deterministic without mocking 'pg' module
// eslint-disable-next-line @typescript-eslint/ban-types
type PoolFactory = (options: { connectionString: string; max: number; idleTimeoutMillis: number }) => Pool;
let createPool: PoolFactory = (options) => new Pool(options as any);

// Test-only: override how pools are created
export function __setPoolFactory(factory: PoolFactory) {
  createPool = factory;
}

export async function initDb(config: DbConfig): Promise<Pool> {
  if (pool) return pool;
  if (!config?.connectionString) {
    throw new Error('DATABASE_URL (connectionString) is required');
  }

  pool = createPool({
    connectionString: config.connectionString,
    max: config.max ?? 10,
    idleTimeoutMillis: config.idleTimeoutMillis ?? 30000,
  });

  // surface errors (do not crash process in tests)
  pool.on('error', (err) => {
    // eslint-disable-next-line no-console
    console.error('pg pool error', err);
  });

  return pool;
}

export function getDb(): Pool {
  if (!pool) throw new Error('DB not initialized. Call initDb() first.');
  return pool;
}

export async function closeDb(): Promise<void> {
  if (pool) {
    const p = pool;
    pool = null;
    await p.end();
  }
}

export async function query(sql: string, params?: any[]): Promise<QueryResult<any>> {
  const p = getDb();
  return p.query(sql, params as any);
}

export async function transaction<T>(callback: (client: PoolClient) => Promise<T>): Promise<T> {
  const p = getDb();
  const client = await p.connect();
  try {
    await client.query('BEGIN');
    const result = await callback(client);
    await client.query('COMMIT');
    return result;
  } catch (err) {
    try {
      await client.query('ROLLBACK');
    } catch {
      // ignore rollback errors
    }
    throw err;
  } finally {
    client.release();
  }
}