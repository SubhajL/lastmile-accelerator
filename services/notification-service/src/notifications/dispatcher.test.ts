import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { NotificationJob } from '../consumers/types.js';
import { createDispatcher, type ChannelAdapter, type ChannelRegistry, type Metrics } from './dispatcher.js';

function job(overrides: Partial<NotificationJob> = {}): NotificationJob {
  return {
    id: 'job-1',
    tenantId: 'tenant-1',
    userId: 'user-1',
    channel: 'email',
    templateName: 'snapshot-ready',
    priority: 'normal',
    payload: { snapshotId: 'snap-123' },
    createdAt: new Date().toISOString(),
    attempt: 0,
    maxAttempts: 3,
    ...overrides
  };
}

describe('notifications/dispatcher', () => {
  let queue: any;
  let metrics: Metrics;
  let registry: ChannelRegistry;
  let adapter: ChannelAdapter;

  beforeEach(() => {
    queue = {
      dequeue: vi.fn().mockResolvedValue([]),
      ack: vi.fn().mockResolvedValue(undefined),
      nack: vi.fn().mockResolvedValue(undefined)
    };
    metrics = { increment: vi.fn() };
    adapter = { send: vi.fn().mockResolvedValue({ ok: true }) };
    registry = { get: vi.fn().mockReturnValue(adapter) };
  });

  it('processNextBatch acks on successful send', async () => {
    queue.dequeue.mockResolvedValue([job({ id: 'a' })]);
    const d = createDispatcher({ queue, registry, metrics, now: () => Date.now() });

    const summary = await d.processNextBatch(5);

    expect(summary.processed).toBe(1);
    expect(queue.ack).toHaveBeenCalledWith('a');
    expect(metrics.increment).toHaveBeenCalledWith('notify_sent', { channel: 'email' });
  });

  it('processNextBatch nacks with retry on adapter failure', async () => {
    adapter.send = vi.fn().mockResolvedValue({ ok: false, error: 'SMTP error' });
    queue.dequeue.mockResolvedValue([job({ id: 'b' })]);
    const d = createDispatcher({ queue, registry, metrics, now: () => Date.now() });

    const summary = await d.processNextBatch(5);

    expect(summary.failed).toBe(1);
    expect(queue.nack).toHaveBeenCalledWith('b', 'SMTP error', true);
    expect(metrics.increment).toHaveBeenCalledWith('notify_failed', { channel: 'email', reason: 'adapter_error' });
  });

  it('processNextBatch nacks without retry when no adapter', async () => {
    registry.get = vi.fn().mockReturnValue(undefined);
    queue.dequeue.mockResolvedValue([job({ id: 'c', channel: 'unknown' as any })]);

    const d = createDispatcher({ queue, registry, metrics, now: () => Date.now() });
    const summary = await d.processNextBatch(5);

    expect(summary.failed).toBe(1);
    expect(queue.nack).toHaveBeenCalledWith('c', 'No adapter for channel unknown', false);
    expect(metrics.increment).toHaveBeenCalledWith('notify_failed', { channel: 'unknown', reason: 'no_channel' });
  });

  it('handles multiple jobs in one batch', async () => {
    queue.dequeue.mockResolvedValue([job({ id: 'd' }), job({ id: 'e' })]);

    const d = createDispatcher({ queue, registry, metrics, now: () => Date.now() });
    const summary = await d.processNextBatch(2);

    expect(summary.processed).toBe(2);
    expect(queue.ack).toHaveBeenCalledTimes(2);
  });
});
