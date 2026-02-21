import type { NotificationJob } from '../../consumers/types.js';
import { withTimeout, createRetry, TimeoutError } from '../reliability.js';

export interface TwilioClient {
  messages: { create: (args: { to: string; from: string; body: string }) => Promise<unknown> };
}

export function createTwilioSmsChannel(opts: {
  client: TwilioClient;
  from: string;
  resolveTo: (job: NotificationJob) => Promise<string>;
  renderTemplate: (template: string, payload: Record<string, unknown>) => Promise<{ subject?: string; text?: string; html?: string }>;
  metrics: { increment: (name: string, labels?: Record<string, string | number>) => void };
  reliability?: { timeoutMs: number; retry: { max: number; baseMs: number; jitterPct: number }; sleep?: (ms: number) => Promise<void> };
}) {
  return {
    async send(job: NotificationJob): Promise<{ ok: true } | { ok: false; error: string }> {
      try {
        const to = await opts.resolveTo(job);
        const { subject, text } = await opts.renderTemplate(job.templateName, job.payload);
        const body = text || subject || '[no content]';
        const retry = opts.reliability && createRetry<void>({
          max: opts.reliability.retry.max,
          baseMs: opts.reliability.retry.baseMs,
          jitterPct: opts.reliability.retry.jitterPct,
          shouldRetry: (e) => e instanceof TimeoutError,
          sleep: opts.reliability.sleep ?? ((ms) => new Promise((r) => setTimeout(r, ms)))
        });
        const sendOnce = async () => { await opts.client.messages.create({ to, from: opts.from, body }); };
        
        if (opts.reliability) {
          await retry!(async () => withTimeout(() => sendOnce(), opts.reliability!.timeoutMs));
        } else {
          await sendOnce();
        }
        opts.metrics.increment('notify_sent', { channel: 'sms' });
        return { ok: true };
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e);
        opts.metrics.increment('notify_failed', { channel: 'sms', reason: 'adapter_error' });
        return { ok: false, error: msg };
      }
    }
  };
}
