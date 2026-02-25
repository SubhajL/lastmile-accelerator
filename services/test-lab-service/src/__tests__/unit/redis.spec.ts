import { describe, expect, test, beforeEach, vi } from 'vitest';
import type { RedisWrapper } from '../../clients/redis.js';

type MockRedisClient = {
  connect: ReturnType<typeof vi.fn>;
  quit: ReturnType<typeof vi.fn>;
  on: ReturnType<typeof vi.fn>;
  get: ReturnType<typeof vi.fn>;
  set: ReturnType<typeof vi.fn>;
  setEx: ReturnType<typeof vi.fn>;
  del: ReturnType<typeof vi.fn>;
  incr: ReturnType<typeof vi.fn>;
  expire: ReturnType<typeof vi.fn>;
};

const mockStore = new Map<string, { value: string; expiresAt?: number }>();
let onErrorHandler: ((err: unknown) => void) | undefined;

const mockClient: MockRedisClient = {
  connect: vi.fn().mockResolvedValue(undefined),
  quit: vi.fn().mockResolvedValue(undefined),
  on: vi.fn((event: string, handler: (err: unknown) => void) => {
    if (event === 'error') onErrorHandler = handler;
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
