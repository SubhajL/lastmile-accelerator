import { describe, it, expect, vi } from 'vitest';
import { createRuntime } from './runtime.js';

describe('bootstrap/runtime nats metrics', () => {
  it('increments nats_processed and nats_failed per subject', async () => {
    const queue = { dequeue: vi.fn().mockResolvedValue([]), ack: vi.fn(), nack: vi.fn() };
    const metrics = { increment: vi.fn() };
    const subscribe = vi.fn().mockImplementation(({ subject }) => ({ once: vi.fn().mockResolvedValue(subject === 'a' ? { processed: 2, failed: 0 } : { processed: 0, failed: 1 }) }));

    const rt = createRuntime({
      queue: queue as any,
      subscribe: subscribe as any,
      dispatcherFactory: undefined,
      metrics,
      now: () => Date.now(),
      subjects: ['a', 'b'],
      batchSize: 1,
      email: { from: 'n', transporter: {} as any, renderTemplate: vi.fn(), resolveTo: vi.fn() },
      defaultMaxAttempts: 3
    });

    await rt.startOnce();

    expect(metrics.increment).toHaveBeenCalledWith('nats_processed', { subject: 'a' });
    expect(metrics.increment).toHaveBeenCalledWith('nats_failed', { subject: 'b' });
  });
});
