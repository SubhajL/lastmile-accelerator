import pg from 'pg';
import type { PostgresConfig } from '../types.js';

const { Pool } = pg;

export function createDbClient(config: PostgresConfig): pg.Pool {
  const pool = new Pool({
    host: config.host,
    port: config.port,
    database: config.database,
    user: config.user,
    password: config.password,
    max: config.maxConnections,
    idleTimeoutMillis: 30000,
    connectionTimeoutMillis: 5000
  });

  pool.on('error', (err) => {
    console.error('Unexpected database error', err);
  });

  return pool;
}

export async function healthCheck(pool: pg.Pool): Promise<boolean> {
  try {
    await pool.query('SELECT 1 as result');
    return true;
  } catch (error) {
    console.error('Database health check failed', error);
    return false;
  }
}
