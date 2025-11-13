import { createHmac } from 'crypto';
import type { NotificationJob } from '../../consumers/types.js';
import type { ChannelAdapter } from '../../notifications/dispatcher.js';
import { withTimeout, createRetry, TimeoutError } from '../reliability.js';

export function createWebhookChannel(opts: {
  http: (url: string, init: RequestInit) => Promise<{ ok: boolean; status: number }>;
  url: string;
  signingSecret?: string;
  metrics: { increment: (name: string, labels?: Record<string, string | number>) => void };
  breaker: { execute: <T>(fn: () => Promise<T>) => Promise<T> };
  reliability: { timeoutMs: number; retry: { max: number; baseMs: number; jitterPct: number }; sleep?: (ms: number) => Promise<void> };
}): ChannelAdapter {
  const retry = createRetry<{ ok: true } | never>({
    max: opts.reliability.retry.max,
    baseMs: opts.reliability.retry.baseMs,
    jitterPct: opts.reliability.retry.jitterPct,
    shouldRetry: (e) => {
      if (e instanceof TimeoutError) return true;
      const maybe = e as { retryable?: unknown };
      return typeof maybe?.retryable === 'boolean' && maybe.retryable === true;
    },
    sleep: opts.reliability.sleep ?? ((ms) => new Promise((r) => setTimeout(r, ms)))
  });

  return {
    async send(job: NotificationJob) {
      const payload = JSON.stringify({ template: job.templateName, payload: job.payload });
      const headers: Record<string, string> = { 'Content-Type': 'application/json' };
      if (opts.signingSecret) {
        const sig = createHmac('sha256', opts.signingSecret).update(payload).digest('hex');
        headers['X-Signature'] = sig;
      }

      try {
        await opts.breaker.execute(async () =>
          retry(async () => {
            const res = await withTimeout(() => opts.http(opts.url, { method: 'POST', headers, body: payload } as RequestInit), opts.reliability.timeoutMs);
            if (!res.ok) {
              const err: Error & { retryable?: boolean } = new Error(`HTTP ${res.status}`);
              err.retryable = res.status >= 500;
              throw err;
            }
            return { ok: true as const };
          })
        );
        opts.metrics.increment('notify_sent', { channel: 'webhook' });
        return { ok: true as const };
      } catch (e) {
        const msg = e instanceof TimeoutError ? 'Operation timed out' : e instanceof Error ? e.message : String(e);
        opts.metrics.increment('notify_failed', { channel: 'webhook', reason: e instanceof TimeoutError ? 'timeout' : 'adapter_error' });
        return { ok: false as const, error: msg };
      }
    }
  };
}
