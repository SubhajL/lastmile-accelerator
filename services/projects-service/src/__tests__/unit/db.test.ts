import { describe, it, expect, beforeEach, vi } from 'vitest';
import { initDb, getDb, closeDb, query, transaction, __setPoolFactory } from '../../db';

// Local test doubles
class MockClient {
  public queries: Array<{ sql: string; params?: any[] }> = [];
  release = vi.fn();
  query = vi.fn(async (sql: string, params?: any[]) => {
    this.queries.push({ sql, params });
    return { rows: [], rowCount: 0 } as any;
  });
}
class MockPool {
  public ended = false;
  public client: MockClient = new MockClient();
  public options: any;
  constructor(options: any) {
    this.options = options;
  }
  on = vi.fn();
  query = vi.fn(async () => {
    return { rows: [{ ok: true }], rowCount: 1 } as any;
  });
  connect = vi.fn(async () => this.client);
  end = vi.fn(async () => {
    this.ended = true;
  });
}

beforeEach(() => {
  __setPoolFactory((opts: any) => new (MockPool as any)(opts));
});

describe('db.ts', () => {
  beforeEach(async () => {
    await closeDb().catch(() => {});
  });

  it('initDb should create a singleton Pool with connectionString', async () => {
    const conn = 'postgres://localhost:5432/appdb';
    const pool = await initDb({ connectionString: conn, max: 5 });

    expect(pool).toBeDefined();
    const anyPool = pool as unknown as { options: any };
    expect(anyPool.options.connectionString ?? anyPool.options?.connectionString).toBe(conn);

    // Subsequent calls return same instance
    const pool2 = getDb();
    expect(pool2).toBe(pool);
  });

  it('query should delegate to pool.query and return rows', async () => {
    await initDb({ connectionString: 'postgres://localhost:5432/db' });
    const res = await query('select 1 as ok');

    expect(res.rowCount).toBe(1);
    expect((res as any).rows[0].ok).toBe(true);
    const pool = getDb() as any;
    expect(pool.query).toHaveBeenCalledWith('select 1 as ok', undefined);
  });

  it('transaction should BEGIN, run callback, and COMMIT on success', async () => {
    await initDb({ connectionString: 'postgres://localhost:5432/db' });

    const result = await transaction(async (client: any) => {
      // perform arbitrary query inside transaction
      await client.query('insert into t values ($1)', [42]);
      return 'done';
    });

    expect(result).toBe('done');

    const pool = getDb() as any;
    const client = pool.client;
    expect(client.query).toHaveBeenNthCalledWith(1, 'BEGIN');
    expect(client.query).toHaveBeenCalledWith('insert into t values ($1)', [42]);
    expect(client.query).toHaveBeenCalledWith('COMMIT');
    expect(client.release).toHaveBeenCalled();
  });

  it('transaction should ROLLBACK and rethrow on error', async () => {
    await initDb({ connectionString: 'postgres://localhost:5432/db' });

    await expect(
      transaction(async () => {
        throw new Error('boom');
      })
    ).rejects.toThrow('boom');

    const pool = getDb() as any;
    const client = pool.client;
    expect(client.query).toHaveBeenNthCalledWith(1, 'BEGIN');
    expect(client.query).toHaveBeenCalledWith('ROLLBACK');
    expect(client.release).toHaveBeenCalled();
  });

  it('closeDb should end the pool and reset singleton', async () => {
    await initDb({ connectionString: 'postgres://localhost:5432/db' });
    const pool = getDb() as any;

    await closeDb();

    expect(pool.end).toHaveBeenCalled();
    expect(() => getDb()).toThrowError();
  });
});