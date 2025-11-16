import { describe, it, expect, vi } from 'vitest';
import * as api from '@opentelemetry/api';
import { createSlackChannel } from './slack.js';

function tracerSpies() {
  const span = { setAttributes: vi.fn(), setStatus: vi.fn(), end: vi.fn(), recordException: vi.fn() } as any;
  const tracer = { startSpan: vi.fn().mockReturnValue(span) } as any;
  vi.spyOn(api.trace, 'getTracer').mockReturnValue(tracer);
  return { tracer, span };
}

describe('slack channel tracing', () => {
  it('emits span with url and channel attrs', async () => {
    const { tracer, span } = tracerSpies();
    const ch = createSlackChannel({
      http: vi.fn().mockResolvedValue({ ok: true, status: 200 }),
      webhookUrl: 'https://hooks.slack.com/abc',
      metrics: { increment: vi.fn() },
      breaker: { execute: (fn: any) => fn() },
      reliability: { timeoutMs: 1000, retry: { max: 1, baseMs: 1, jitterPct: 0 } },
    });
    const r = await ch.send({
      id: '1', tenantId: 't', userId: 'u', channel: 'slack', templateName: 't', priority: 'normal', payload: {}, createdAt: new Date().toISOString(), attempt: 0, maxAttempts: 3,
    } as any);
    expect(r).toEqual({ ok: true });
    expect(tracer.startSpan).toHaveBeenCalledWith('channel.slack.send', expect.any(Object));
    expect(span.setAttributes).toHaveBeenCalledWith(expect.objectContaining({ channel: 'slack', url: 'https://hooks.slack.com/abc' }));
  });
});
