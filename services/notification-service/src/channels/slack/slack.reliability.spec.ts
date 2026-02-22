import { describe, it, expect, vi } from 'vitest';
import { createSlackChannel } from './slack';
import { createCircuitBreaker } from '../circuitbreaker';

function job(): any { return { templateName: 'snapshot-ready', payload: { id: 's1' } }; }

describe('slack reliability', () => {
  it('retries on timeout and 5xx, not 4xx', async () => {
    let calls = 0;
    const http = vi.fn().mockImplementation(async (_url: string, init: RequestInit) => {
      expect((init as any).signal).toBeDefined();
      calls += 1;
      if (calls === 1) { await new Promise((r) => setTimeout(r, 50)); return { ok: true, status: 200 }; }
      if (calls === 2) { return { ok: false, status: 500 }; }
      return { ok: true, status: 200 };
    });
    const breaker = createCircuitBreaker({ failureThreshold: 10, halfOpenAfterMs: 1000, windowSize: 10, now: () => Date.now() });
    const ch = createSlackChannel({ http: http as any, webhookUrl: 'https://hooks.slack.com/xxx', metrics: { increment: vi.fn() }, breaker, reliability: { timeoutMs: 5, retry: { max: 3, baseMs: 1, jitterPct: 0 } } });
    const res = await ch.send(job());
    expect(res.ok).toBe(true);
    expect(calls).toBe(3);

    // 4xx should not retry
    const http4xx = vi.fn().mockImplementation(async (_url: string, init: RequestInit) => {
      expect((init as any).signal).toBeDefined();
      return { ok: false, status: 400 };
    });
    const ch2 = createSlackChannel({ http: http4xx as any, webhookUrl: 'https://hooks.slack.com/xxx', metrics: { increment: vi.fn() }, breaker, reliability: { timeoutMs: 50, retry: { max: 3, baseMs: 1, jitterPct: 0 } } });
    const res2 = await ch2.send(job());
    expect(res2.ok).toBe(false);
    expect(http4xx).toHaveBeenCalledTimes(1);
  });
});
