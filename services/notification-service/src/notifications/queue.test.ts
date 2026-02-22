import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { createNotificationQueue } from './queue.js';
import type { NotificationJob } from '../consumers/types.js';
import type Redis from 'ioredis';

describe('notifications/queue', () => {
  let mockRedis: Redis;
  const mockJob: Omit<NotificationJob, 'id' | 'createdAt' | 'attempt'> = {
    tenantId: 'tenant-1',
    userId: 'user-1',
    channel: 'email',
    templateName: 'snapshot-ready',
    priority: 'normal',
    payload: { snapshotId: 'snap-123' },
    maxAttempts: 3
  };

  beforeEach(() => {
    mockRedis = {
      zadd: vi.fn().mockResolvedValue(1),
      zrange: vi.fn().mockResolvedValue([]),
      zrem: vi.fn().mockResolvedValue(1),
      zcard: vi.fn().mockResolvedValue(0),
      hset: vi.fn().mockResolvedValue(1),
      hget: vi.fn().mockResolvedValue(null),
      hdel: vi.fn().mockResolvedValue(1),
      hlen: vi.fn().mockResolvedValue(0),
      eval: vi.fn().mockResolvedValue([]),
      pipeline: vi.fn().mockReturnValue({
        zadd: vi.fn().mockReturnThis(),
        hset: vi.fn().mockReturnThis(),
        zrem: vi.fn().mockReturnThis(),
        exec: vi.fn().mockResolvedValue([[null, 1], [null, 1]])
      })
    } as unknown as Redis;
  });

  describe('createNotificationQueue', () => {
    it('should create notification queue with Redis client', () => {
      const queue = createNotificationQueue(mockRedis);

      expect(queue).toBeDefined();
      expect(typeof queue.enqueue).toBe('function');
      expect(typeof queue.dequeue).toBe('function');
    });
  });

  describe('enqueue', () => {
    it('should enqueue notification with priority', async () => {
      const queue = createNotificationQueue(mockRedis);

      const jobId = await queue.enqueue(mockJob);

      expect(jobId).toBeDefined();
      expect(typeof jobId).toBe('string');
      expect(mockRedis.pipeline).toHaveBeenCalled();
    });

    it('should assign higher score to critical priority', async () => {
      const queue = createNotificationQueue(mockRedis);
      const criticalJob = { ...mockJob, priority: 'critical' as const };

      await queue.enqueue(criticalJob);

      const pipeline = (mockRedis.pipeline as unknown as ReturnType<typeof vi.fn>).mock.results[0]!
        .value as { zadd: ReturnType<typeof vi.fn> };
      expect(pipeline.zadd).toHaveBeenCalledWith(
        'notifications:queue',
        expect.any(Number),
        expect.any(String)
      );
    });
  });

  describe('dequeue', () => {
    it('should dequeue notifications in priority order', async () => {
      const jobData = JSON.stringify({
        ...mockJob,
        id: 'job-1',
        createdAt: new Date().toISOString(),
        attempt: 0
      });

      (mockRedis as any).eval = vi.fn().mockResolvedValue([jobData]);

      const queue = createNotificationQueue(mockRedis);
      const jobs = await queue.dequeue(1);

      expect(jobs).toHaveLength(1);
      expect(jobs[0].id).toBe('job-1');
      expect((mockRedis as any).eval).toHaveBeenCalledWith(
        expect.any(String),
        3,
        'notifications:queue',
        'notifications:jobs',
        'notifications:processing',
        1,
        expect.any(String),
      );
    });

    it('should return empty array when queue is empty', async () => {
      (mockRedis as any).eval = vi.fn().mockResolvedValue([]);

      const queue = createNotificationQueue(mockRedis);
      const jobs = await queue.dequeue(5);

      expect(jobs).toHaveLength(0);
    });

    it('should move jobs to processing set', async () => {
      const jobData = JSON.stringify({
        ...mockJob,
        id: 'job-1',
        createdAt: new Date().toISOString(),
        attempt: 0
      });

      (mockRedis as any).eval = vi.fn().mockResolvedValue([jobData]);

      const queue = createNotificationQueue(mockRedis);
      await queue.dequeue(1);

      expect((mockRedis as any).eval).toHaveBeenCalled();
    });
  });

  describe('ack', () => {
    it('should acknowledge successfully processed notification', async () => {
      const queue = createNotificationQueue(mockRedis);

      await queue.ack('job-1');

      expect(mockRedis.pipeline).toHaveBeenCalled();
    });
  });

  describe('nack', () => {
    it('should nack and requeue failed notification with retry', async () => {
      const jobData = JSON.stringify({
        ...mockJob,
        id: 'job-1',
        createdAt: new Date().toISOString(),
        attempt: 0
      });

      mockRedis.hget = vi.fn().mockResolvedValue(jobData);

      const queue = createNotificationQueue(mockRedis);
      await queue.nack('job-1', 'Test error', true);

      expect(mockRedis.hget).toHaveBeenCalledWith('notifications:jobs', 'job-1');
      expect(mockRedis.pipeline).toHaveBeenCalled();
    });

    it('should move to DLQ after max retries exceeded', async () => {
      const jobData = JSON.stringify({
        ...mockJob,
        id: 'job-1',
        createdAt: new Date().toISOString(),
        attempt: 3,
        maxAttempts: 3
      });

      mockRedis.hget = vi.fn().mockResolvedValue(jobData);

      const queue = createNotificationQueue(mockRedis);
      await queue.nack('job-1', 'Max retries', true);

      expect(mockRedis.pipeline).toHaveBeenCalled();
    });

    it('should not requeue when retry is false', async () => {
      const jobData = JSON.stringify({
        ...mockJob,
        id: 'job-1',
        createdAt: new Date().toISOString(),
        attempt: 0
      });

      mockRedis.hget = vi.fn().mockResolvedValue(jobData);

      const queue = createNotificationQueue(mockRedis);
      await queue.nack('job-1', 'Permanent error', false);

      expect(mockRedis.pipeline).toHaveBeenCalled();
    });
  });

  describe('getQueueDepth', () => {
    it('should return accurate queue depth', async () => {
      mockRedis.zcard = vi.fn().mockResolvedValue(5);

      const queue = createNotificationQueue(mockRedis);
      const depth = await queue.getQueueDepth();

      expect(depth).toBe(5);
      expect(mockRedis.zcard).toHaveBeenCalledWith('notifications:queue');
    });
  });

  describe('getProcessingCount', () => {
    it('should return accurate processing count', async () => {
      mockRedis.hlen = vi.fn().mockResolvedValue(3);

      const queue = createNotificationQueue(mockRedis);
      const count = await queue.getProcessingCount();

      expect(count).toBe(3);
      expect(mockRedis.hlen).toHaveBeenCalledWith('notifications:processing');
    });
  });
});
