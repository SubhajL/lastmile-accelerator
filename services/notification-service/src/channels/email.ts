import type nodemailer from 'nodemailer';
import type { NotificationJob } from '../consumers/types.js';
import { withTimeout, createRetry, TimeoutError } from './reliability.js';

export type TemplateRenderer = (
  templateName: string,
  payload: Record<string, unknown>
) => Promise<{ subject: string; html: string; text?: string }>;

export interface EmailChannelOptions {
  transporter: Pick<nodemailer.Transporter, 'sendMail'>;
  renderTemplate: TemplateRenderer;
  from: string;
  resolveTo: (job: NotificationJob) => Promise<string>;
  reliability?: { timeoutMs: number; retry: { max: number; baseMs: number; jitterPct: number }; sleep?: (ms: number) => Promise<void> };
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
        const retry = opts.reliability && createRetry<void>({
          max: opts.reliability.retry.max,
          baseMs: opts.reliability.retry.baseMs,
          jitterPct: opts.reliability.retry.jitterPct,
          shouldRetry: (e) => e instanceof TimeoutError,
          sleep: opts.reliability.sleep ?? ((ms) => new Promise((r) => setTimeout(r, ms)))
        });
        const sendOnce = async () => opts.transporter.sendMail({ from: opts.from, to, subject, html, text });
        if (opts.reliability) {
          await retry!(async () => withTimeout(() => sendOnce(), opts.reliability!.timeoutMs));
        } else {
          await sendOnce();
        }
        return { ok: true };
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        return { ok: false, error: msg };
      }
    }
  };
}
