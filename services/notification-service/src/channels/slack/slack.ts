import type { NotificationJob } from '../../consumers/types.js';
import type { ChannelAdapter } from '../../notifications/dispatcher.js';
import { withTimeout, createRetry, TimeoutError } from '../reliability.js';

export function createSlackChannel(opts: {
  http: (url: string, init: RequestInit) => Promise<{ ok: boolean; status: number }>;
  webhookUrl: string;
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
      const text = `${job.templateName}: ${JSON.stringify(job.payload)}`;
      try {
        await opts.breaker.execute(async () =>
          retry(async () => {
            const res = await withTimeout(() => opts.http(opts.webhookUrl, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ text }) } as RequestInit), opts.reliability.timeoutMs);
            if (!res.ok) {
              const err: Error & { retryable?: boolean } = new Error(`HTTP ${res.status}`);
              err.retryable = res.status >= 500;
              throw err;
            }
            return { ok: true as const };
          })
        );
        opts.metrics.increment('notify_sent', { channel: 'slack' });
        return { ok: true as const };
      } catch (e) {
        const msg = e instanceof TimeoutError ? 'Operation timed out' : e instanceof Error ? e.message : String(e);
        opts.metrics.increment('notify_failed', { channel: 'slack', reason: e instanceof TimeoutError ? 'timeout' : 'adapter_error' });
        return { ok: false as const, error: msg };
      }
    }
  };
}
