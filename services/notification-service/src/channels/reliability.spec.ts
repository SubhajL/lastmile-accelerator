import { describe, expect, test, vi } from 'vitest';

import { TimeoutError, withTimeout, createRetry } from './reliability';

function sleep(ms: number) { return new Promise((r) => setTimeout(r, ms)); }

describe('reliability helpers', () => {
  test('withTimeout returns result under budget', async () => {
    vi.useFakeTimers();
    const p = withTimeout(async () => {
      await sleep(5);
      return 42;
    }, 50);
    await vi.advanceTimersByTimeAsync(5);
    await expect(p).resolves.toEqual(42);
    vi.useRealTimers();
  });

  test('withTimeout rejects TimeoutError after ms', async () => {
    vi.useFakeTimers();
    const p = withTimeout(async () => {
      await sleep(50);
      return 'x';
    }, 5);
    const assertion = expect(p).rejects.toBeInstanceOf(TimeoutError);
    await vi.advanceTimersByTimeAsync(5);
    await assertion;
    vi.useRealTimers();
  });

  test('retry stops on success before max', async () => {
    let calls = 0;
    const retry = createRetry({ max: 3, baseMs: 1, jitterPct: 0, shouldRetry: () => true, sleep: async () => {} });
    const result = await retry(async () => {
      calls += 1;
      if (calls < 2) throw new Error('flaky');
      return 'ok';
    });
    expect(result).toEqual('ok');
    expect(calls).toEqual(2);
  });

  test('retry stops on non-retryable error', async () => {
    const retry = createRetry({ max: 3, baseMs: 1, jitterPct: 0, shouldRetry: (e) => (e as any).retryable === true, sleep: async () => {} });
    await expect(retry(async () => { throw Object.assign(new Error('boom'), { retryable: false }); })).rejects.toBeInstanceOf(Error);
  });

  test('backoff increases exponentially with jitter bounds', async () => {
    const delays: number[] = [];
    const sleepMock = vi.fn(async (ms: number) => { delays.push(ms); });
    let attempts = 0;
    const retry = createRetry({ max: 3, baseMs: 10, jitterPct: 0.2, shouldRetry: () => true, sleep: sleepMock });
    await expect(retry(async () => { attempts += 1; throw new Error('always'); })).rejects.toBeInstanceOf(Error);
    // There are 3 attempts; 2 sleeps between them (after attempt 1 and 2)
    expect(delays.length).toBe(2);
    // Each delay should be around 10, 20 with ±20% jitter
    const within = (ms: number, target: number) => ms >= target * 0.8 && ms <= target * 1.2;
    expect(within(delays[0]!, 10)).toBe(true);
    expect(within(delays[1]!, 20)).toBe(true);
    expect(attempts).toBe(3);
  });
});
