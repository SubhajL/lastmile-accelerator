import type { NotificationJob } from '../../consumers/types.js';

export interface TwilioClient {
  messages: { create: (args: { to: string; from: string; body: string }) => Promise<unknown> };
}

export function createTwilioSmsChannel(opts: {
  client: TwilioClient;
  from: string;
  resolveTo: (job: NotificationJob) => Promise<string>;
  renderTemplate: (template: string, payload: Record<string, unknown>) => Promise<{ subject?: string; text?: string; html?: string }>;
  metrics: { increment: (name: string, labels?: Record<string, string | number>) => void };
}) {
  return {
    async send(job: NotificationJob): Promise<{ ok: true } | { ok: false; error: string }> {
      try {
        const to = await opts.resolveTo(job);
const { subject, text } = await opts.renderTemplate(job.templateName, job.payload);
        const body = text || subject || '[no content]';
        await opts.client.messages.create({ to, from: opts.from, body });
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
