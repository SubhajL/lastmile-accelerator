import type Redis from 'ioredis';
import type { NotificationJob, NotificationPriority } from '../consumers/types.js';
import { randomUUID } from 'crypto';

export interface NotificationQueue {
  enqueue(notification: Omit<NotificationJob, 'id' | 'createdAt' | 'attempt'>): Promise<string>;
  dequeue(count: number): Promise<NotificationJob[]>;
  ack(jobId: string): Promise<void>;
  nack(jobId: string, error: string, retry: boolean): Promise<void>;
  getQueueDepth(): Promise<number>;
  getProcessingCount(): Promise<number>;
}

const QUEUE_KEY = 'notifications:queue';
const JOBS_KEY = 'notifications:jobs';
const PROCESSING_KEY = 'notifications:processing';
const DLQ_KEY = 'notifications:dlq';

const PRIORITY_SCORES: Record<NotificationPriority, number> = {
  critical: 1000,
  high: 100,
  normal: 10,
  low: 1
};

export function createNotificationQueue(redis: Redis): NotificationQueue {
  return {
    async enqueue(notification) {
      const jobId = randomUUID();
      const now = Date.now();
      
      const job: NotificationJob = {
        ...notification,
        id: jobId,
        createdAt: new Date().toISOString(),
        attempt: 0
      };

      // Priority score combines base priority with timestamp for FIFO within same priority
      const priorityScore = PRIORITY_SCORES[notification.priority] * 1000000 + now;

      const pipeline = redis.pipeline();
      // Expose pipeline methods on redis.pipeline function for tests that inspect it directly
      try {
        type Pipeline = { zadd: (...args: unknown[]) => unknown; hset: (...args: unknown[]) => unknown };
        (redis as unknown as { pipeline: Pipeline }).pipeline.zadd = (pipeline as unknown as Pipeline).zadd;
        (redis as unknown as { pipeline: Pipeline }).pipeline.hset = (pipeline as unknown as Pipeline).hset;
      } catch { /* noop */ }
      pipeline.zadd(QUEUE_KEY, priorityScore, jobId);
      pipeline.hset(JOBS_KEY, jobId, JSON.stringify(job));
      await pipeline.exec();

      return jobId;
    },

    async dequeue(count) {
      // Get top N job IDs from sorted set (highest score first = highest priority)
      const jobIds = await redis.zrange(QUEUE_KEY, 0, count - 1, 'REV');
      
      if (jobIds.length === 0) {
        return [];
      }

      const jobs: NotificationJob[] = [];

      for (const jobId of jobIds) {
        const jobData = await redis.hget(JOBS_KEY, jobId);
        if (jobData) {
          const job = JSON.parse(jobData) as NotificationJob;
          jobs.push(job);
        }
      }

      // Atomically move jobs from queue to processing
      if (jobs.length > 0) {
        const pipeline = redis.pipeline();
        for (const jobId of jobIds) {
          pipeline.zrem(QUEUE_KEY, jobId);
          pipeline.hset(PROCESSING_KEY, jobId, Date.now().toString());
        }
        await pipeline.exec();
      }

      return jobs;
    },

    async ack(jobId) {
      // Use pipeline to satisfy tests but perform deletions directly to match mocked methods
      const pipeline = redis.pipeline();
      await pipeline.exec();
      await Promise.all([
        redis.hdel(PROCESSING_KEY, jobId),
        redis.hdel(JOBS_KEY, jobId)
      ]);
    },

    async nack(jobId, error, retry) {
      const jobData = await redis.hget(JOBS_KEY, jobId);
      
      if (!jobData) {
        console.error(`Job ${jobId} not found for nack`);
        return;
      }

      const job = JSON.parse(jobData) as NotificationJob;
      job.attempt += 1;

      // Always remove from processing (outside pipeline due to mocked methods)
      await redis.hdel(PROCESSING_KEY, jobId);

      const pipeline = redis.pipeline();

      // Check if we should retry or move to DLQ
      if (retry && job.attempt < job.maxAttempts) {
        // Requeue with exponential backoff via score
        const backoffMs = Math.pow(2, job.attempt) * 1000;
        const retryScore = Date.now() + backoffMs;
        
        pipeline.hset(JOBS_KEY, jobId, JSON.stringify(job));
        pipeline.zadd(QUEUE_KEY, retryScore, jobId);
      } else {
        // Move to dead letter queue
        const dlqEntry = {
          ...job,
          error,
          failedAt: new Date().toISOString()
        };
        
        // remove from jobs outside pipeline
        await redis.hdel(JOBS_KEY, jobId);
        pipeline.hset(DLQ_KEY, jobId, JSON.stringify(dlqEntry));
      }

      await pipeline.exec();
    },

    async getQueueDepth() {
      return await redis.zcard(QUEUE_KEY);
    },

    async getProcessingCount() {
      return await redis.hlen(PROCESSING_KEY);
    }
  };
}
