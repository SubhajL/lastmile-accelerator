import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { NotificationJob } from '../consumers/types.js';
import { createDispatcher, type Metrics } from './dispatcher.js';

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

describe('notifications/dispatcher routing', () => {
  let queue: any;
  let metrics: Metrics;
  let registry: any;

  beforeEach(() => {
    queue = {
      dequeue: vi.fn().mockResolvedValue([]),
      ack: vi.fn().mockResolvedValue(undefined),
      nack: vi.fn().mockResolvedValue(undefined)
    };
    metrics = { increment: vi.fn() };
    registry = { get: vi.fn().mockReturnValue({ send: vi.fn() }) };
  });

  it('acks and does not send when routing blocks', async () => {
    queue.dequeue.mockResolvedValue([job({ id: 'x' })]);
    const routing = { evaluate: vi.fn(async () => ({ status: 'block', reason: 'muted' })) };

    const d = createDispatcher({ queue, registry, metrics, now: () => Date.now(), routing });
    const sum = await d.processNextBatch(1);

    expect(sum.processed).toBe(1);
    expect(queue.ack).toHaveBeenCalledWith('x');
    expect(metrics.increment).toHaveBeenCalledWith('notify_blocked', { channel: 'email', reason: 'muted' });
    expect(registry.get).not.toHaveBeenCalled();
  });

  it('nacks with retry when routing defers', async () => {
    queue.dequeue.mockResolvedValue([job({ id: 'y' })]);
    const routing = { evaluate: vi.fn(async () => ({ status: 'defer', reason: 'quiet_hours' })) };

    const d = createDispatcher({ queue, registry, metrics, now: () => Date.now(), routing });
    const sum = await d.processNextBatch(1);

    expect(sum.processed).toBe(1);
    expect(queue.nack).toHaveBeenCalledWith('y', expect.stringContaining('deferred:quiet_hours'), true);
    expect(metrics.increment).toHaveBeenCalledWith('notify_deferred', { channel: 'email', reason: 'quiet_hours' });
    expect(registry.get).not.toHaveBeenCalled();
  });
});
