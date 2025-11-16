import { describe, it, expect, vi } from 'vitest';
import * as api from '@opentelemetry/api';
import { createWebhookChannel } from './http.js';

function tracerSpies() {
  const span = { setAttributes: vi.fn(), setStatus: vi.fn(), end: vi.fn(), recordException: vi.fn() } as any;
  const tracer = { startSpan: vi.fn().mockReturnValue(span) } as any;
  vi.spyOn(api.trace, 'getTracer').mockReturnValue(tracer);
  return { tracer, span };
}

describe('webhook channel tracing', () => {
  it('emits span with url and channel attrs', async () => {
    const { tracer, span } = tracerSpies();
    const ch = createWebhookChannel({
      http: vi.fn().mockResolvedValue({ ok: true, status: 200 }),
      url: 'https://example.com/hook',
      metrics: { increment: vi.fn() },
      breaker: { execute: (fn: any) => fn() },
      reliability: { timeoutMs: 1000, retry: { max: 1, baseMs: 1, jitterPct: 0 } },
    });
    const r = await ch.send({
      id: '1', tenantId: 't', userId: 'u', channel: 'webhook', templateName: 't', priority: 'normal', payload: {}, createdAt: new Date().toISOString(), attempt: 0, maxAttempts: 3,
    } as any);
    expect(r).toEqual({ ok: true });
    expect(tracer.startSpan).toHaveBeenCalledWith('channel.webhook.send', expect.any(Object));
    expect(span.setAttributes).toHaveBeenCalledWith(expect.objectContaining({ channel: 'webhook', url: 'https://example.com/hook' }));
  });
});
