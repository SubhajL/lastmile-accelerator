import { createClient } from 'redis';
import type { RedisClientType } from 'redis';

export interface RedisClientLogger {
  error: (obj: unknown, msg?: string) => void;
}

export interface CreateRedisClientOptions {
  logger?: RedisClientLogger;
}

export interface RedisWrapper {
  get: (key: string) => Promise<string | null>;
  set: (key: string, value: string, ttlSec?: number) => Promise<void>;
  del: (key: string) => Promise<void>;
  incr: (key: string) => Promise<number>;
  incrWithExpire: (key: string, ttlSec: number) => Promise<number>;
  expire: (key: string, ttlSec: number) => Promise<void>;
  close: () => Promise<void>;
}

/**
 * Create a Redis client wrapper with typed operations.
 * Provides error handling, connection lifecycle, and common operations.
 *
 * @param url - Redis connection URL (e.g., redis://localhost:6379)
 * @returns RedisWrapper with typed operations
 */
export async function createRedisClient(
  url: string,
  opts: CreateRedisClientOptions = {}
): Promise<RedisWrapper> {
  const client: RedisClientType = createClient({ url });

  // Handle connection errors
  client.on('error', (err) => {
    opts.logger?.error({ err }, 'Redis client error');
  });

  // Connect to Redis
  await client.connect();

  /**
   * Get value from Redis by key.
   * Returns null if key doesn't exist.
   */
  async function get(key: string): Promise<string | null> {
    return await client.get(key);
  }

  /**
   * Set value in Redis with optional TTL.
   * @param key - Redis key
   * @param value - String value to store
   * @param ttlSec - Optional TTL in seconds
   */
  async function set(key: string, value: string, ttlSec?: number): Promise<void> {
    if (ttlSec !== undefined && ttlSec > 0) {
      await client.setEx(key, ttlSec, value);
    } else {
      await client.set(key, value);
    }
  }

  /**
   * Delete key from Redis.
   */
  async function del(key: string): Promise<void> {
    await client.del(key);
  }

  /**
   * Atomically increment counter.
   * Creates key with value 1 if it doesn't exist.
   * @returns New value after increment
   */
  async function incr(key: string): Promise<number> {
    return await client.incr(key);
  }

  /**
   * Atomically increment counter and set TTL.
   * Prevents race where INCR succeeds but EXPIRE is never applied.
   *
   * @returns New value after increment
   */
  async function incrWithExpire(key: string, ttlSec: number): Promise<number> {
    const res = await client.multi().incr(key).expire(key, ttlSec).exec();
    if (!res) throw new Error('Redis transaction aborted');
    const newValue = res[0];
    if (typeof newValue !== 'number') throw new Error('Redis INCR returned non-number result');
    return newValue;
  }

  /**
   * Set TTL on existing key.
   * @param key - Redis key
   * @param ttlSec - TTL in seconds
   */
  async function expire(key: string, ttlSec: number): Promise<void> {
    await client.expire(key, ttlSec);
  }

  /**
   * Gracefully close Redis connection.
   * Waits for pending operations to complete.
   */
  async function close(): Promise<void> {
    await client.quit();
  }

  return {
    get,
    set,
    del,
    incr,
    incrWithExpire,
    expire,
    close,
  };
}
