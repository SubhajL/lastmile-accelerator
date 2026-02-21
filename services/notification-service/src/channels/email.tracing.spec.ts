import { describe, it, expect, vi } from 'vitest';
import * as api from '@opentelemetry/api';
import { createEmailChannel } from './email.js';

function tracerSpies() {
  const span = { setAttributes: vi.fn(), setStatus: vi.fn(), end: vi.fn(), recordException: vi.fn() } as any;
  const tracer = { startSpan: vi.fn().mockReturnValue(span) } as any;
  vi.spyOn(api.trace, 'getTracer').mockReturnValue(tracer);
  return { tracer, span };
}

describe('email channel tracing', () => {
  it('emits span with to and channel attrs', async () => {
    const { tracer, span } = tracerSpies();
    const ch = createEmailChannel({
      transporter: { sendMail: vi.fn().mockResolvedValue({}) } as any,
      renderTemplate: vi.fn().mockResolvedValue({ subject: 's', html: '<p/>', text: 't' }),
      from: 'noreply@example.com',
      resolveTo: vi.fn().mockResolvedValue('user@example.com'),
    });
    const r = await ch.send({
      id: '1', tenantId: 't', userId: 'u', channel: 'email', templateName: 't', priority: 'normal', payload: {}, createdAt: new Date().toISOString(), attempt: 0, maxAttempts: 3,
    } as any);
    expect(r).toEqual({ ok: true });
    expect(tracer.startSpan).toHaveBeenCalledWith('channel.email.send', expect.any(Object));
    expect(span.setAttributes).toHaveBeenCalledWith(expect.objectContaining({ channel: 'email', to: 'user@example.com' }));
  });
});
