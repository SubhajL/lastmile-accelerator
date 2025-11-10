import { describe, it, expect, vi } from 'vitest';
import { createRunLoop } from './runloop.js';

function makeDispatcher() {
  return { processNextBatch: vi.fn().mockResolvedValue({ processed: 1, failed: 0 }) };
}

describe('bootstrap/runloop', () => {
  it('start triggers periodic processNextBatch calls', async () => {
    vi.useFakeTimers();
    const dispatcher = makeDispatcher();
    const rl = createRunLoop({ dispatcher: dispatcher as any, batchSize: 5, tickMs: 200, logger: console });

    rl.start();
    await vi.advanceTimersByTimeAsync(200);
    await Promise.resolve();
    await vi.advanceTimersByTimeAsync(200);
    await Promise.resolve();
    await vi.advanceTimersByTimeAsync(200);
    await Promise.resolve();

    expect(dispatcher.processNextBatch).toHaveBeenCalledTimes(3);
    expect(dispatcher.processNextBatch).toHaveBeenCalledWith(5);

    rl.stop();
    vi.useRealTimers();
  });

  it('stop halts future ticks', async () => {
    vi.useFakeTimers();
    const dispatcher = makeDispatcher();
    const rl = createRunLoop({ dispatcher: dispatcher as any, batchSize: 2, tickMs: 100, logger: console });

    rl.start();
    vi.advanceTimersByTime(250);
    rl.stop();
    const calls = (dispatcher.processNextBatch as any).mock.calls.length;

    vi.advanceTimersByTime(500);

    expect((dispatcher.processNextBatch as any).mock.calls.length).toBe(calls);
    vi.useRealTimers();
  });

  it('prevents overlapping ticks when previous still running', async () => {
    vi.useFakeTimers();
    const dispatcher = { processNextBatch: vi.fn(() => new Promise(res => setTimeout(() => res({ processed: 1, failed: 0 }), 300))) };
    const rl = createRunLoop({ dispatcher: dispatcher as any, batchSize: 1, tickMs: 100, logger: console });

    rl.start();
    await vi.advanceTimersByTimeAsync(350);

    // Only one call should be in-flight/finished despite 3 ticks
    expect(dispatcher.processNextBatch).toHaveBeenCalledTimes(1);

    // Let the promise resolve and allow another tick
    await vi.advanceTimersByTimeAsync(300);
    await Promise.resolve();
    expect(dispatcher.processNextBatch).toHaveBeenCalledTimes(2);

    rl.stop();
    vi.useRealTimers();
  });
});
