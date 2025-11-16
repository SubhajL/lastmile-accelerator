import { describe, it, expect, vi } from 'vitest';
import * as api from '@opentelemetry/api';
import { upsertOutboxPending } from './pg-outbox.js';

describe('pg-outbox tracing', () => {
  it('wraps upsert in span with key attribute', async () => {
    const fakeSpan = makeFakeSpan();
    const fakeTracer = { startSpan: vi.fn().mockReturnValue(fakeSpan) } as unknown as api.Tracer;
    vi.spyOn(api.trace, 'getTracer').mockReturnValue(fakeTracer);

    const db = { query: vi.fn().mockResolvedValue({ rows: [{ ok: 1 }] }) } as any;
    const inserted = await upsertOutboxPending(db, 'k1');
    expect(inserted).toBe(true);
    expect(fakeTracer.startSpan).toHaveBeenCalledWith('outbox.pg.upsertPending', expect.any(Object));
    expect(fakeSpan.setAttributes).toHaveBeenCalledWith({ key: 'k1' });
    expect(fakeSpan.setStatus).toHaveBeenCalledWith({ code: api.SpanStatusCode.OK });
    expect(fakeSpan.end).toHaveBeenCalled();
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
