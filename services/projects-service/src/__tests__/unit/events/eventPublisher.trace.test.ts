import { describe, it, expect, vi, beforeEach } from 'vitest';
import * as events from '../../../events/eventPublisher';
import * as nats from '../../../nats';
import * as otApi from '@opentelemetry/api';

function fakeSpan() {
  return {
    spanContext() { return { traceId: 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', spanId: 'bbbbbbbbbbbbbbbb', traceFlags: 1 }; },
  } as any;
}

describe('events/eventPublisher traceparent propagation', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('injects active traceparent when not provided', async () => {
    vi.spyOn(otApi.trace, 'getActiveSpan').mockReturnValue(fakeSpan());
    const pub = vi.spyOn(nats, 'publish').mockResolvedValue();
    await events.publishProjectEvent('created', { tenantId: 't1', id: 'p1' });
    const traceHeader = pub.mock.calls[0][2];
    expect(traceHeader).toBe('00-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-bbbbbbbbbbbbbbbb-01');
  });

  it('preserves explicit traceparent argument', async () => {
    const pub = vi.spyOn(nats, 'publish').mockResolvedValue();
    await events.publishMemberEvent('added', { tenantId: 't1', id: 'm1' }, '00-deadbeefdeadbeefdeadbeefdeadbeef-feedfeedfeedfeed-00');
    expect(pub.mock.calls[0][2]).toBe('00-deadbeefdeadbeefdeadbeefdeadbeef-feedfeedfeedfeed-00');
  });
});
