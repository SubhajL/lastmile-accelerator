import { describe, expect, test, beforeEach, vi } from 'vitest';
import type { RedisWrapper } from '../../clients/redis.js';

const mockStore = new Map<string, { value: string; expiresAt?: number }>();
let onErrorHandler: ((err: unknown) => void) | undefined;

const mockClient = {
  connect: vi.fn().mockResolvedValue(undefined),
  quit: vi.fn().mockResolvedValue(undefined),
  on: vi.fn((event: string, handler: (err: unknown) => void) => {
    if (event === 'error') onErrorHandler = handler;
  }),
  multi: vi.fn(() => {
    const commands: Array<() => Promise<unknown>> = [];

    const chain = {
      incr(key: string) {
        commands.push(async () => mockClient.incr(key));
        return chain;
      },
      expire(key: string, ttlSec: number) {
        commands.push(async () => mockClient.expire(key, ttlSec));
        return chain;
      },
      async exec() {
        const results: unknown[] = [];
        for (const cmd of commands) results.push(await cmd());
        return results;
      },
    };

    return chain;
  }),
  get: vi.fn(async (key: string) => {
    const entry = mockStore.get(key);
    if (!entry) return null;
    if (entry.expiresAt && Date.now() > entry.expiresAt) {
      mockStore.delete(key);
      return null;
    }
    return entry.value;
  }),
  set: vi.fn(async (key: string, value: string) => {
    mockStore.set(key, { value });
  }),
  setEx: vi.fn(async (key: string, ttlSec: number, value: string) => {
    mockStore.set(key, {
      value,
      expiresAt: Date.now() + ttlSec * 1000,
    });
  }),
  del: vi.fn(async (key: string) => {
    mockStore.delete(key);
  }),
  incr: vi.fn(async (key: string) => {
    const entry = mockStore.get(key);
    const currentValue = entry ? parseInt(entry.value, 10) : 0;
    const newValue = currentValue + 1;
    mockStore.set(key, { value: String(newValue) });
    return newValue;
  }),
  expire: vi.fn(async (key: string, ttlSec: number) => {
    const entry = mockStore.get(key);
    if (entry) {
      entry.expiresAt = Date.now() + ttlSec * 1000;
    }
  }),
};

const createClient = vi.fn(() => mockClient);

vi.mock('redis', () => ({
  createClient,
}));

describe('createRedisClient', () => {
  beforeEach(async () => {
    mockStore.clear();
    onErrorHandler = undefined;
    vi.clearAllMocks();
  });

  test('connects successfully with valid URL', async () => {
    const { createRedisClient } = await import('../../clients/redis.js');
    const redis = await createRedisClient('redis://localhost:6379');

    expect(redis).toBeDefined();
    expect(redis.get).toBeInstanceOf(Function);
    expect(redis.set).toBeInstanceOf(Function);
    expect(redis.close).toBeInstanceOf(Function);
  });

  test('get returns null for non-existent key', async () => {
    const { createRedisClient } = await import('../../clients/redis.js');
    const redis = await createRedisClient('redis://localhost:6379');

    const result = await redis.get('nonexistent-key');
    expect(result).toBeNull();
  });

  test('get returns cached value after set', async () => {
    const { createRedisClient } = await import('../../clients/redis.js');
    const redis = await createRedisClient('redis://localhost:6379');

    const testKey = 'test:cached';
    const testValue = 'test-value-123';

    await redis.set(testKey, testValue);
    const result = await redis.get(testKey);

    expect(result).toBe(testValue);
  });

  test('set stores value with TTL', async () => {
    const { createRedisClient } = await import('../../clients/redis.js');
    const redis = await createRedisClient('redis://localhost:6379');

    const testKey = 'test:ttl';
    const testValue = 'expiring-value';

    await redis.set(testKey, testValue, 60);

    const result = await redis.get(testKey);
    expect(result).toBe(testValue);
  });

  test('set falls back to non-TTL set when ttlSec is 0', async () => {
    const { createRedisClient } = await import('../../clients/redis.js');
    const redis = await createRedisClient('redis://localhost:6379');

    await redis.set('test:ttl:zero', 'x', 0);

    expect(mockClient.set).toHaveBeenCalledWith('test:ttl:zero', 'x');
    expect(mockClient.setEx).not.toHaveBeenCalled();
  });

  test('incr increments counter atomically', async () => {
    const { createRedisClient } = await import('../../clients/redis.js');
    const redis = await createRedisClient('redis://localhost:6379');

    const testKey = 'test:counter';

    const firstIncr = await redis.incr(testKey);
    expect(firstIncr).toBe(1);

    const secondIncr = await redis.incr(testKey);
    expect(secondIncr).toBe(2);

    const thirdIncr = await redis.incr(testKey);
    expect(thirdIncr).toBe(3);
  });

  test('incrWithExpire increments counter and applies TTL', async () => {
    const { createRedisClient } = await import('../../clients/redis.js');
    const redis = await createRedisClient('redis://localhost:6379');

    const testKey = 'test:counter:ttl';
    const now = Date.now();
    const dateNowSpy = vi.spyOn(Date, 'now');
    dateNowSpy.mockReturnValue(now);

    const firstIncr = await redis.incrWithExpire(testKey, 1);
    expect(firstIncr).toBe(1);

    dateNowSpy.mockReturnValue(now + 1500);
    const expired = await redis.get(testKey);
    expect(expired).toBeNull();
    dateNowSpy.mockRestore();
  });

  test('del removes key from cache', async () => {
    const { createRedisClient } = await import('../../clients/redis.js');
    const redis = await createRedisClient('redis://localhost:6379');

    const testKey = 'test:delete';
    const testValue = 'to-be-deleted';

    await redis.set(testKey, testValue);

    const beforeDelete = await redis.get(testKey);
    expect(beforeDelete).toBe(testValue);

    await redis.del(testKey);

    const afterDelete = await redis.get(testKey);
    expect(afterDelete).toBeNull();
  });

  test('expire sets TTL on existing key', async () => {
    const { createRedisClient } = await import('../../clients/redis.js');
    const redis = await createRedisClient('redis://localhost:6379');

    const testKey = 'test:expire';
    const testValue = 'will-expire';

    await redis.set(testKey, testValue);
    await redis.expire(testKey, 60);

    const result = await redis.get(testKey);
    expect(result).toBe(testValue);
  });

  test('close disconnects gracefully', async () => {
    const { createRedisClient } = await import('../../clients/redis.js');
    const redis = await createRedisClient('redis://localhost:6379');

    await expect(redis.close()).resolves.not.toThrow();
  });

  test('logs Redis errors via provided logger', async () => {
    const logger = { error: vi.fn() };
    const { createRedisClient } = await import('../../clients/redis.js');
    await createRedisClient('redis://localhost:6379', { logger });

    const err = new Error('boom');
    onErrorHandler?.(err);
    expect(logger.error).toHaveBeenCalledWith({ err }, 'Redis client error');
  });
});
