import { describe, it, expect, vi } from 'vitest';
import * as api from '@opentelemetry/api';
import { createDispatcher } from './dispatcher.js';

function job(id = 'j1') {
  return {
    id,
    tenantId: 't1',
    userId: 'u1',
    channel: 'email',
    templateName: 'tmpl',
    priority: 'normal',
    payload: {},
    createdAt: new Date().toISOString(),
    attempt: 0,
    maxAttempts: 3,
  } as any;
}

function tracerSpies() {
  const span = { setAttributes: vi.fn(), setStatus: vi.fn(), end: vi.fn(), recordException: vi.fn() } as any;
  const tracer = { startSpan: vi.fn().mockReturnValue(span) } as any;
  vi.spyOn(api.trace, 'getTracer').mockReturnValue(tracer);
  return { tracer, span };
}

describe('dispatcher tracing', () => {
  it('wraps processJob in a span with attrs and OK', async () => {
    const { tracer, span } = tracerSpies();
    const d = createDispatcher({
      queue: { dequeue: vi.fn(), ack: vi.fn(), nack: vi.fn() } as any,
      registry: { get: () => ({ send: vi.fn().mockResolvedValue({ ok: true as const }) }) } as any,
      metrics: { increment: vi.fn() },
      now: Date.now,
    });
    const r = await (d as any).processJob(job('j-ok'));
    expect(r).toBe('ok');
    expect(tracer.startSpan).toHaveBeenCalledWith('dispatch.processJob', expect.any(Object));
    expect(span.setAttributes).toHaveBeenCalledWith(expect.objectContaining({ jobId: 'j-ok', channel: 'email' }));
  });

  it('sets ERROR on adapter failure', async () => {
    const { span } = tracerSpies();
    const d = createDispatcher({
      queue: { dequeue: vi.fn(), ack: vi.fn(), nack: vi.fn() } as any,
      registry: { get: () => ({ send: vi.fn().mockResolvedValue({ ok: false as const, error: 'bad' }) }) } as any,
      metrics: { increment: vi.fn() },
      now: Date.now,
    });
    const r = await (d as any).processJob(job('j-bad'));
    expect(r).toBe('failed');
    expect(span.setStatus).toHaveBeenCalled();
  });
});
