import { describe, it, expect, vi } from 'vitest';
import { createRuntime } from './runtime.js';

function makeQueue() {
  return { dequeue: vi.fn().mockResolvedValue([]), ack: vi.fn(), nack: vi.fn() };
}

function makeRegistry() {
  return { get: vi.fn() };
}

describe('bootstrap/runtime metrics', () => {
  it('increments metrics on success and failure', async () => {
    const queue = makeQueue();
    const metrics = { increment: vi.fn() };
    const registry = { get: vi.fn().mockReturnValue({ send: vi.fn().mockResolvedValue({ ok: true }) }) };
    const now = () => Date.now();
    const rt = (createRuntime({
      queue: { ...queue } as any,
      registry: registry as any,
      metrics,
      now,
      // minimal subscribe/subject for metrics test; not used by processNextBatch
      subscribe: (() => ({ once: vi.fn().mockResolvedValue({ processed: 0, failed: 0 }) })) as any,
      subjects: ['snapshots']
    } as any)) as any;

    // Inject jobs by stubbing dequeue result
    (queue.dequeue as any).mockResolvedValueOnce([{ id: 'j1', channel: 'email' }]);
    await (rt as any).processNextBatch(1);

    expect(metrics.increment).toHaveBeenCalledWith('notify_sent', { channel: 'email' });

    // Failure path
    registry.get = vi.fn().mockReturnValue({ send: vi.fn().mockResolvedValue({ ok: false, error: 'x' }) });
    (queue.dequeue as any).mockResolvedValueOnce([{ id: 'j2', channel: 'email' }]);
    await (rt as any).processNextBatch(1);

    expect(metrics.increment).toHaveBeenCalledWith('notify_failed', { channel: 'email', reason: 'adapter_error' });
  });
});
