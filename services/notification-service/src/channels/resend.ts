import type { NotificationJob } from '../consumers/types.js';

export interface ResendChannelOptions {
  http: (url: string, init: RequestInit) => Promise<{ ok: boolean; status: number; json: () => Promise<any> }>;
  apiKey: string;
  from: string;
  resolveTo: (job: NotificationJob) => Promise<string>;
  renderTemplate: (template: string, payload: Record<string, unknown>) => Promise<{ subject: string; html: string; text?: string }>;
}

export function createResendChannel(opts: ResendChannelOptions) {
  return {
    async send(job: NotificationJob): Promise<{ ok: true } | { ok: false; error: string }> {
      try {
        const to = await opts.resolveTo(job);
        const { subject, html, text } = await opts.renderTemplate(job.templateName, job.payload);
        const body = {
          from: opts.from,
          to,
          subject,
          html,
          text
        };
        const res = await opts.http('https://api.resend.com/emails', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${opts.apiKey}`
          },
          body: JSON.stringify(body)
        } as RequestInit);
        if (!res.ok) return { ok: false, error: `Resend HTTP ${res.status}` };
        return { ok: true };
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e);
        return { ok: false, error: msg };
      }
    }
  };
}
