import type nodemailer from 'nodemailer';
import type { NotificationJob } from '../consumers/types.js';
import { withTimeout, createRetry, TimeoutError } from './reliability.js';
import { getTracer } from '../telemetry/tracing.js';
import { SpanStatusCode } from '@opentelemetry/api';

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
      const tracer = getTracer('notification-service');
      const span = tracer.startSpan('channel.email.send', { attributes: { channel: 'email' } });
      span.setAttributes({ channel: 'email' });
      try {
        const to = await opts.resolveTo(job);
        span.setAttributes({ channel: 'email', to });
        const { subject, html, text } = await opts.renderTemplate(job.templateName, job.payload);
        const reliability = opts.reliability;
        const retry = reliability && createRetry<void>({
          max: reliability.retry.max,
          baseMs: reliability.retry.baseMs,
          jitterPct: reliability.retry.jitterPct,
          shouldRetry: (e) => e instanceof TimeoutError,
          sleep: reliability.sleep ?? ((ms) => new Promise((r) => setTimeout(r, ms)))
        });
        const sendOnce = async () => {
          await opts.transporter.sendMail({ from: opts.from, to, subject, html, text });
        };
        if (reliability) {
          await retry!(async () =>
            withTimeout((signal) => {
              if (signal.aborted) throw new TimeoutError();
              return sendOnce();
            }, reliability.timeoutMs),
          );
        } else {
          await sendOnce();
        }
        span.setStatus({ code: SpanStatusCode.OK });
        span.end();
        return { ok: true };
      } catch (err) {
        const msg = err instanceof Error ? err.message : String(err);
        if (err instanceof Error && typeof (span as { recordException?: (er: Error) => void }).recordException === 'function') {
          (span as { recordException: (er: Error) => void }).recordException(err);
        }
        span.setStatus({ code: SpanStatusCode.ERROR, message: msg });
        span.end();
        return { ok: false, error: msg };
      }
    }
  };
}
