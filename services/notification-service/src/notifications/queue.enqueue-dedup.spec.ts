import { describe, it, expect, vi, beforeEach } from 'vitest';
import type Redis from 'ioredis';
import { createNotificationQueue } from './queue.js';

describe('queue enqueue-time dedup (PG outbox, FF gated)', () => {
  let mockRedis: Redis;
  let db: { query: ReturnType<typeof vi.fn> };

  const base = {
    tenantId: 't1',
    userId: 'u1',
    channel: 'email',
    templateName: 'welcome',
    priority: 'normal' as const,
    payload: { a: 1, b: 'x' },
    maxAttempts: 3
  };

  beforeEach(() => {
    mockRedis = {
      zadd: vi.fn(),
      zrange: vi.fn(),
      zrem: vi.fn(),
      zcard: vi.fn(),
      hset: vi.fn(),
      hget: vi.fn(),
      hdel: vi.fn(),
      hlen: vi.fn(),
      pipeline: vi.fn().mockReturnValue({
        zadd: vi.fn().mockReturnThis(),
        hset: vi.fn().mockReturnThis(),
        zrem: vi.fn().mockReturnThis(),
        exec: vi.fn().mockResolvedValue([[null, 1], [null, 1]])
      })
    } as unknown as Redis;
    db = { query: vi.fn() } as any;
  });

  it('dedup ON enqueues first, suppresses duplicate second', async () => {
    // first insert returns a row → inserted = true; second insert returns no rows → duplicate
    db.query
      .mockResolvedValueOnce({ rows: [{ inserted: 1 }] })
      .mockResolvedValueOnce({ rows: [] });

    const queue = createNotificationQueue(mockRedis, {
      db: db as any,
      outboxDedupEnabled: true,
      dedupKeyFn: () => 'k::t1:u1:email:welcome:{"a":1,"b":"x"}'
    } as any);

    await queue.enqueue(base);
    await queue.enqueue(base);

    // Only one Redis pipeline used since second enqueue suppressed
    expect((mockRedis.pipeline as any)).toHaveBeenCalledTimes(1);
    expect(db.query).toHaveBeenCalledTimes(2);
  });

  it('dedup OFF enqueues duplicates normally', async () => {
    const queue = createNotificationQueue(mockRedis);

    await queue.enqueue(base);
    await queue.enqueue(base);

    expect((mockRedis.pipeline as any)).toHaveBeenCalledTimes(2);
  });

  it('dedup ON enqueues when key is new', async () => {
    db.query.mockResolvedValueOnce({ rows: [{ inserted: 1 }] });

    const queue = createNotificationQueue(mockRedis, {
      db: db as any,
      outboxDedupEnabled: true,
      dedupKeyFn: () => 'k1'
    } as any);

    await queue.enqueue(base);
    expect((mockRedis.pipeline as any)).toHaveBeenCalledTimes(1);
  });
});
