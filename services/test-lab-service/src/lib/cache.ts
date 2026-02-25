import type { RedisWrapper } from '../clients/redis.js';

/**
 * TTL constants for common caching patterns
 */
export const TTL = {
  SHORT: 300, // 5 minutes
  MEDIUM: 1800, // 30 minutes
  LONG: 3600, // 1 hour
} as const;

export interface CacheHelper {
  cacheAside: <T>(key: string, fetchFn: () => Promise<T>, ttlSec: number) => Promise<T>;
  set: (key: string, value: unknown, ttlSec?: number) => Promise<void>;
  get: <T>(key: string) => Promise<T | null>;
  invalidate: (key: string) => Promise<void>;
}

/**
 * Create a cache helper with common caching patterns.
 * All keys are automatically prefixed to avoid collisions.
 *
 * @param redis - Redis client wrapper
 * @param prefix - Key prefix (e.g., 'test-lab-service:')
 * @returns CacheHelper instance
 */
export function createCacheHelper(redis: RedisWrapper, prefix: string): CacheHelper {
  /**
   * Build fully-qualified cache key with prefix
   */
  function buildKey(key: string): string {
    return `${prefix}${key}`;
  }

  /**
   * Cache-aside pattern: check cache → fetch on miss → populate cache
   *
   * Handles cache errors gracefully by falling back to fetchFn.
   * This ensures availability even when Redis fails.
   *
   * @param key - Cache key (without prefix)
   * @param fetchFn - Function to fetch data on cache miss
   * @param ttlSec - TTL in seconds
   * @returns Cached or freshly fetched data
   */
  async function cacheAside<T>(key: string, fetchFn: () => Promise<T>, ttlSec: number): Promise<T> {
    const fullKey = buildKey(key);

    try {
      // Try to get from cache
      const cached = await redis.get(fullKey);

      if (cached !== null) {
        // Cache hit - parse and return
        return JSON.parse(cached) as T;
      }
    } catch (error) {
      // Cache read error - log and continue to fetch
      console.error(`Cache read error for key ${fullKey}:`, error);
    }

    // Cache miss or error - fetch fresh data
    const freshData = await fetchFn();

    try {
      // Try to populate cache for next time
      await redis.set(fullKey, JSON.stringify(freshData), ttlSec);
    } catch (error) {
      // Cache write error - log but don't fail the request
      console.error(`Cache write error for key ${fullKey}:`, error);
    }

    return freshData;
  }

  /**
   * Direct cache set with JSON serialization
   *
   * @param key - Cache key (without prefix)
   * @param value - Value to cache (will be JSON serialized)
   * @param ttlSec - Optional TTL in seconds
   */
  async function set(key: string, value: unknown, ttlSec?: number): Promise<void> {
    const fullKey = buildKey(key);
    await redis.set(fullKey, JSON.stringify(value), ttlSec);
  }

  /**
   * Direct cache get with JSON deserialization
   *
   * @param key - Cache key (without prefix)
   * @returns Cached value or null if not found or corrupted
   */
  async function get<T>(key: string): Promise<T | null> {
    const fullKey = buildKey(key);
    const cached = await redis.get(fullKey);

    if (cached === null) {
      return null;
    }

    try {
      return JSON.parse(cached) as T;
    } catch (error) {
      // Corrupted cache entry - log and return null
      console.error(`Cache parse error for key ${fullKey}:`, error);
      // Optionally delete corrupted entry
      await redis.del(fullKey).catch(() => {});
      return null;
    }
  }

  /**
   * Invalidate single cache key
   *
   * Note: This implementation supports exact key deletion only.
   * For pattern-based invalidation (e.g., wildcard matching),
   * implement SCAN-based deletion to avoid blocking Redis.
   *
   * @param key - Cache key to delete (without prefix)
   */
  async function invalidate(key: string): Promise<void> {
    const fullKey = buildKey(key);
    await redis.del(fullKey);
  }

  return {
    cacheAside,
    set,
    get,
    invalidate,
  };
}
