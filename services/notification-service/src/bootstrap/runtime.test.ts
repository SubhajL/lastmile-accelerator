import { describe, it, expect, vi } from 'vitest';
import { createRuntime, startNatsSubscriptions } from './runtime.js';

function makeSubscribeOnce(processed = 1, failed = 0) {
  return () => ({ once: vi.fn().mockResolvedValue({ processed, failed }) });
}

describe('bootstrap/runtime', () => {
  it('wires subscription and dispatcher, processes a batch', async () => {
    const queue = { dequeue: vi.fn(), ack: vi.fn(), nack: vi.fn() };
    const metrics = { increment: vi.fn() };
    const now = () => Date.now();

    const dispatcher = { processNextBatch: vi.fn().mockResolvedValue({ processed: 2, failed: 0 }) };
    const dispatcherFactory = vi.fn().mockReturnValue(dispatcher);

    const subscribe = vi.fn().mockImplementation(makeSubscribeOnce(1, 0));

    const resolveTo = vi.fn().mockResolvedValue('user@example.com');
    const transporter = { sendMail: vi.fn() };
    const renderTemplate = vi.fn().mockResolvedValue({ subject: 'x', html: 'y', text: 'y' });

    const rt = createRuntime({
      queue: queue as any,
      subscribe,
      dispatcherFactory,
      metrics: metrics as any,
      now,
      subjects: ['snapshot.*', 'fixes.*'],
      batchSize: 10,
      email: {
        from: 'noreply@example.com',
        transporter: transporter as any,
        renderTemplate,
        resolveTo
      },
      defaultMaxAttempts: 3
    });

    const summary = await rt.startOnce();

    expect(subscribe).toHaveBeenCalled();
    expect(dispatcher.processNextBatch).toHaveBeenCalledWith(10);
    expect(summary).toEqual({ processed: 2, failed: 0 });
  });

  it('startNatsSubscriptions catches rejections (no unhandledRejection)', async () => {
    let unhandled = false;
    const onUnhandled = () => {
      unhandled = true;
    };

    process.once('unhandledRejection', onUnhandled);

    const subscribe = vi.fn().mockReturnValue({
      once: vi.fn().mockRejectedValue(new Error('boom')),
    });

    const router = { route: vi.fn() };
    const handle = await startNatsSubscriptions(['a'], subscribe as any, router as any);

    await new Promise((r) => setTimeout(r, 0));
    await handle.stop();
    await new Promise((r) => setTimeout(r, 0));

    process.removeListener('unhandledRejection', onUnhandled);
    expect(unhandled).toBe(false);
  });
});
