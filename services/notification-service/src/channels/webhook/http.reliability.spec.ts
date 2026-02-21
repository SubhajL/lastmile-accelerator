import { describe, it, expect, vi } from 'vitest';
import { createWebhookChannel } from '../webhook/http.js';
import { createCircuitBreaker } from '../circuitbreaker.js';
import { TimeoutError } from '../reliability.js';

function job(): any { return { templateName: 'publish-failed', payload: { snapshotId: 's1', error: 'boom' } }; }

describe('webhook reliability', () => {
  it('times out slow HTTP and retries configured times', async () => {
    let calls = 0;
    const http = vi.fn().mockImplementation(async () => {
      calls += 1;
      await new Promise((r) => setTimeout(r, 50));
      return { ok: true, status: 200 };
    });
    const breaker = createCircuitBreaker({ failureThreshold: 10, halfOpenAfterMs: 1000, windowSize: 10, now: () => Date.now() });
    const ch = createWebhookChannel({
      http: http as any,
      url: 'https://example.com/hook',
      signingSecret: 's',
      metrics: { increment: vi.fn() },
      breaker,
      reliability: { timeoutMs: 5, retry: { max: 3, baseMs: 1, jitterPct: 0 } }
    });

    const res = await ch.send(job());
    expect(res.ok).toBe(false);
    expect((res as any).error).toContain('timed out');
    expect(calls).toBe(3); // 3 attempts
  });

  it('opens circuit after consecutive failures; blocks calls', async () => {
    const http = vi.fn().mockResolvedValue({ ok: false, status: 500 });
    const breaker = createCircuitBreaker({ failureThreshold: 2, halfOpenAfterMs: 10_000, windowSize: 10, now: () => Date.now() });
    const ch = createWebhookChannel({
      http: http as any,
      url: 'https://example.com/hook',
      metrics: { increment: vi.fn() },
      breaker,
      reliability: { timeoutMs: 100, retry: { max: 1, baseMs: 1, jitterPct: 0 } }
    } as any);

    const r1 = await ch.send(job());
    const r2 = await ch.send(job());
    expect(r1.ok).toBe(false);
    expect(r2.ok).toBe(false);

    // On third call, breaker should be open and fail fast (no extra http call)
    const callsBefore = http.mock.calls.length;
    const r3 = await ch.send(job());
    const callsAfter = http.mock.calls.length;
    expect(r3.ok).toBe(false);
    expect((r3 as any).error).toContain('circuit breaker');
    expect(callsAfter).toBe(callsBefore); // no new HTTP attempt
  });
});
