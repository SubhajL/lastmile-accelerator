import { createClient } from 'redis';
import type { RedisClientType } from 'redis';

export interface RedisWrapper {
  get: (key: string) => Promise<string | null>;
  set: (key: string, value: string, ttlSec?: number) => Promise<void>;
  del: (key: string) => Promise<void>;
  incr: (key: string) => Promise<number>;
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
export async function createRedisClient(url: string): Promise<RedisWrapper> {
  const client: RedisClientType = createClient({ url });

  // Handle connection errors
  client.on('error', (err) => {
    console.error('Redis Client Error:', err);
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
    if (ttlSec !== undefined) {
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
    expire,
    close,
  };
}
