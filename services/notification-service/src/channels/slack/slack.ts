import type { NotificationJob } from '../../consumers/types.js';
import type { ChannelAdapter } from '../../notifications/dispatcher.js';

export function createSlackChannel(opts: { http: (url: string, init: RequestInit) => Promise<{ ok: boolean; status: number }>; webhookUrl: string; metrics: { increment: (name: string, labels?: Record<string,string|number>) => void }; breaker: { execute: <T>(fn: () => Promise<T>) => Promise<T> } }): ChannelAdapter {
  return {
    async send(job: NotificationJob) {
      const text = `${job.templateName}: ${JSON.stringify(job.payload)}`;
      const res = await opts.breaker.execute(async () => opts.http(opts.webhookUrl, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ text }) } as RequestInit));
      if (!res.ok) {
        opts.metrics.increment('notify_failed', { channel: 'slack', reason: 'adapter_error' });
        return { ok: false as const, error: `HTTP ${res.status}` };
      }
      opts.metrics.increment('notify_sent', { channel: 'slack' });
      return { ok: true as const };
    }
  };
}
