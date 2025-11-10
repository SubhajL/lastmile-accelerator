import { createHmac } from 'crypto';
import type { NotificationJob } from '../../consumers/types.js';
import type { ChannelAdapter } from '../../notifications/dispatcher.js';

export function createWebhookChannel(opts: { http: (url: string, init: RequestInit) => Promise<{ ok: boolean; status: number }>; url: string; signingSecret?: string; metrics: { increment: (name: string, labels?: Record<string,string|number>) => void }; breaker: { execute: <T>(fn: () => Promise<T>) => Promise<T> } }): ChannelAdapter {
  return {
    async send(job: NotificationJob) {
      const payload = JSON.stringify({ template: job.templateName, payload: job.payload });
      const headers: Record<string,string> = { 'Content-Type': 'application/json' };
      if (opts.signingSecret) {
        const sig = createHmac('sha256', opts.signingSecret).update(payload).digest('hex');
        headers['X-Signature'] = sig;
      }
      const res = await opts.breaker.execute(async () => opts.http(opts.url, { method: 'POST', headers, body: payload } as RequestInit));
      if (!res.ok) {
        opts.metrics.increment('notify_failed', { channel: 'webhook', reason: 'adapter_error' });
        return { ok: false as const, error: `HTTP ${res.status}` };
      }
      opts.metrics.increment('notify_sent', { channel: 'webhook' });
      return { ok: true as const };
    }
  };
}
