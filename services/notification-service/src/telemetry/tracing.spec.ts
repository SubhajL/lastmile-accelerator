import { describe, it, expect, vi } from 'vitest';
import * as api from '@opentelemetry/api';

// Import after vi spies
import { getTracer, withSpan, recordRetry } from './tracing.js';

describe('telemetry/tracing helpers', () => {
  it('withSpan sets attributes, status OK, and ends', async () => {
    const fakeSpan = makeFakeSpan();
    const fakeTracer = { startSpan: vi.fn().mockReturnValue(fakeSpan) } as unknown as api.Tracer;
    vi.spyOn(api.trace, 'getTracer').mockReturnValue(fakeTracer);

    const res = await withSpan('test.span', { a: 1, b: 'x' }, async () => 42);
    expect(res).toBe(42);
    expect(fakeTracer.startSpan).toHaveBeenCalledWith('test.span', expect.any(Object));
    expect(fakeSpan.setAttributes).toHaveBeenCalledWith({ a: 1, b: 'x' });
    expect(fakeSpan.setStatus).toHaveBeenCalledWith({ code: api.SpanStatusCode.OK });
    expect(fakeSpan.end).toHaveBeenCalled();
  });

  it('withSpan records exception and sets ERROR on failure', async () => {
    const fakeSpan = makeFakeSpan();
    const fakeTracer = { startSpan: vi.fn().mockReturnValue(fakeSpan) } as unknown as api.Tracer;
    vi.spyOn(api.trace, 'getTracer').mockReturnValue(fakeTracer);

    await expect(withSpan('fail.span', { k: 'v' }, async () => { throw new Error('boom'); })).rejects.toThrow('boom');
    expect(fakeSpan.recordException).toHaveBeenCalled();
    expect(fakeSpan.setStatus).toHaveBeenCalledWith({ code: api.SpanStatusCode.ERROR, message: expect.stringContaining('boom') });
    expect(fakeSpan.end).toHaveBeenCalled();
  });

  it('recordRetry adds "retry" event with attempt and reason', () => {
    const s = makeFakeSpan();
    recordRetry(s as any, 2, 'timeout');
    expect(s.addEvent).toHaveBeenCalledWith('retry', { attempt: 2, reason: 'timeout' });
  });
});

function makeFakeSpan() {
  return {
    setAttributes: vi.fn(),
    setStatus: vi.fn(),
    recordException: vi.fn(),
    addEvent: vi.fn(),
    end: vi.fn(),
  } as unknown as api.Span;
}
