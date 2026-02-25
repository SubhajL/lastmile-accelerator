import { describe, expect, test, beforeEach, vi } from 'vitest';
import type { RedisWrapper } from '../../clients/redis.js';
import { createCacheHelper, TTL } from '../../lib/cache.js';

describe('createCacheHelper', () => {
  let mockRedis: RedisWrapper;
  const testPrefix = 'test-lab-service:';

  beforeEach(() => {
    const mockStore = new Map<string, string>();

    mockRedis = {
      get: vi.fn(async (key: string) => mockStore.get(key) ?? null),
      set: vi.fn(async (key: string, value: string, ttlSec?: number) => {
        mockStore.set(key, value);
      }),
      del: vi.fn(async (key: string) => {
        mockStore.delete(key);
      }),
      incr: vi.fn(async (key: string) => {
        const current = parseInt(mockStore.get(key) ?? '0', 10);
        const newValue = current + 1;
        mockStore.set(key, String(newValue));
        return newValue;
      }),
      expire: vi.fn(async () => {}),
      close: vi.fn(async () => {}),
    };
  });

  describe('cacheAside', () => {
    test('fetches from function on cache miss', async () => {
      const cache = createCacheHelper(mockRedis, testPrefix);
      const fetchFn = vi.fn(async () => ({ id: '123', name: 'Test' }));

      const result = await cache.cacheAside('user:123', fetchFn, TTL.SHORT);

      expect(result).toEqual({ id: '123', name: 'Test' });
      expect(fetchFn).toHaveBeenCalledTimes(1);
      expect(mockRedis.get).toHaveBeenCalledWith('test-lab-service:user:123');
      expect(mockRedis.set).toHaveBeenCalledWith(
        'test-lab-service:user:123',
        JSON.stringify({ id: '123', name: 'Test' }),
        TTL.SHORT,
      );
    });

    test('returns cached value on hit without calling fetchFn', async () => {
      const cache = createCacheHelper(mockRedis, testPrefix);
      const cachedData = { id: '456', name: 'Cached User' };

      // Pre-populate cache
      await mockRedis.set('test-lab-service:user:456', JSON.stringify(cachedData));

      const fetchFn = vi.fn(async () => ({ id: '456', name: 'Fresh User' }));

      const result = await cache.cacheAside('user:456', fetchFn, TTL.SHORT);

      expect(result).toEqual(cachedData);
      expect(fetchFn).not.toHaveBeenCalled();
      expect(mockRedis.get).toHaveBeenCalledWith('test-lab-service:user:456');
    });

    test('stores fetched value in cache with TTL', async () => {
      const cache = createCacheHelper(mockRedis, testPrefix);
      const fetchFn = async () => ({ data: 'test' });
      const customTTL = 1800;

      await cache.cacheAside('test:key', fetchFn, customTTL);

      expect(mockRedis.set).toHaveBeenCalledWith(
        'test-lab-service:test:key',
        JSON.stringify({ data: 'test' }),
        customTTL,
      );
    });

    test('handles cache errors gracefully without throwing', async () => {
      const faultyRedis: RedisWrapper = {
        get: vi.fn(async () => {
          throw new Error('Redis connection failed');
        }),
        set: vi.fn(async () => {}),
        del: vi.fn(async () => {}),
        incr: vi.fn(async () => 1),
        expire: vi.fn(async () => {}),
        close: vi.fn(async () => {}),
      };

      const cache = createCacheHelper(faultyRedis, testPrefix);
      const fetchFn = vi.fn(async () => ({ fallback: true }));

      const result = await cache.cacheAside('error:key', fetchFn, TTL.SHORT);

      expect(result).toEqual({ fallback: true });
      expect(fetchFn).toHaveBeenCalled();
    });
  });

  describe('set and get', () => {
    test('round-trip serialize/deserialize JSON', async () => {
      const cache = createCacheHelper(mockRedis, testPrefix);
      const complexObject = {
        id: 789,
        name: 'Complex',
        tags: ['a', 'b', 'c'],
        metadata: { created: '2024-01-01' },
      };

      await cache.set('complex:obj', complexObject, TTL.MEDIUM);

      const retrieved = await cache.get<typeof complexObject>('complex:obj');

      expect(retrieved).toEqual(complexObject);
      expect(mockRedis.set).toHaveBeenCalledWith(
        'test-lab-service:complex:obj',
        JSON.stringify(complexObject),
        TTL.MEDIUM,
      );
    });

    test('get returns null for non-existent key', async () => {
      const cache = createCacheHelper(mockRedis, testPrefix);

      const result = await cache.get('nonexistent');

      expect(result).toBeNull();
    });
  });

  describe('invalidate', () => {
    test('deletes single cache key', async () => {
      const cache = createCacheHelper(mockRedis, testPrefix);

      await cache.invalidate('user:123');

      expect(mockRedis.del).toHaveBeenCalledWith('test-lab-service:user:123');
    });
  });

  describe('TTL constants', () => {
    test('SHORT equals 300 seconds', () => {
      expect(TTL.SHORT).toBe(300);
    });

    test('MEDIUM equals 1800 seconds', () => {
      expect(TTL.MEDIUM).toBe(1800);
    });

    test('LONG equals 3600 seconds', () => {
      expect(TTL.LONG).toBe(3600);
    });
  });
});
