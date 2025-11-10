import type nodemailer from 'nodemailer';
import type { NotificationJob } from '../consumers/types.js';

export type TemplateRenderer = (
  templateName: string,
  payload: Record<string, unknown>
) => Promise<{ subject: string; html: string; text?: string }>;

export interface EmailChannelOptions {
  transporter: Pick<nodemailer.Transporter, 'sendMail'>;
  renderTemplate: TemplateRenderer;
  from: string;
  resolveTo: (job: NotificationJob) => Promise<string>;
}

export type SendResult = { ok: true } | { ok: false; error: string };

export interface EmailChannel {
  send(job: NotificationJob): Promise<SendResult>;
}

export function createEmailChannel(opts: EmailChannelOptions): EmailChannel {
  return {
    async send(job: NotificationJob): Promise<SendResult> {
      try {
        const to = await opts.resolveTo(job);
        const { subject, html, text } = await opts.renderTemplate(job.templateName, job.payload);
        await opts.transporter.sendMail({
          from: opts.from,
          to,
          subject,
          html,
          text
        });
        return { ok: true };
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        return { ok: false, error: msg };
      }
    }
  };
}
