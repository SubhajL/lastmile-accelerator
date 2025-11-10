import type { NotificationJob } from '../consumers/types.js';

export interface Metrics {
  increment(name: string, labels?: Record<string, string | number>): void;
}

export interface ChannelAdapter {
  send(job: NotificationJob): Promise<{ ok: true } | { ok: false; error: string }>;
}

export interface ChannelRegistry {
  get(channel: string): ChannelAdapter | undefined;
}

export interface DispatcherOptions {
  queue: {
    dequeue: (count: number) => Promise<NotificationJob[]>;
    ack: (jobId: string) => Promise<void>;
    nack: (jobId: string, error: string, retry: boolean) => Promise<void>;
  };
  registry: ChannelRegistry;
  metrics: Metrics;
  now: () => number;
  routing?: { evaluate: (job: NotificationJob) => Promise<{ status: 'allow' } | { status: 'block'; reason: string } | { status: 'defer'; reason: string }> };
}

export interface DispatchSummary {
  processed: number;
  failed: number;
}

export function createDispatcher(opts: DispatcherOptions) {
  async function processJob(job: NotificationJob): Promise<'ok' | 'failed'> {
    if (opts.routing) {
      const decision = await opts.routing.evaluate(job);
      if (decision.status === 'block') {
        await opts.queue.ack(job.id);
        opts.metrics.increment('notify_blocked', { channel: job.channel, reason: decision.reason });
        return 'ok';
      }
      if (decision.status === 'defer') {
        await opts.queue.nack(job.id, `deferred:${decision.reason}`, true);
        opts.metrics.increment('notify_deferred', { channel: job.channel, reason: decision.reason });
        return 'ok';
      }
    }

    const adapter = opts.registry.get(job.channel);
    if (!adapter) {
      await opts.queue.nack(job.id, `No adapter for channel ${job.channel}`, false);
      opts.metrics.increment('notify_failed', { channel: job.channel, reason: 'no_channel' });
      return 'failed';
    }

    const res = await adapter.send(job);
    if (res.ok) {
      await opts.queue.ack(job.id);
      opts.metrics.increment('notify_sent', { channel: job.channel });
      return 'ok';
    }

    await opts.queue.nack(job.id, res.error, true);
    opts.metrics.increment('notify_failed', { channel: job.channel, reason: 'adapter_error' });
    return 'failed';
  }

  async function processNextBatch(max: number): Promise<DispatchSummary> {
    const jobs = await opts.queue.dequeue(max);
    let processed = 0;
    let failed = 0;

    for (const j of jobs) {
      const r = await processJob(j);
      processed += 1;
      if (r === 'failed') failed += 1;
    }

    return { processed, failed };
  }

  return {
    processJob,
    processNextBatch
  };
}
