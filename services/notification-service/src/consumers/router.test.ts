import { describe, it, expect, vi } from 'vitest';
import { createEventRouter, type EventEnvelope } from './router.js';

function envelope(overrides: Partial<EventEnvelope> = {}): EventEnvelope {
  return {
    type: 'snapshot.ready',
    data: { snapshotId: 'snap-123', tenantId: 'tenant-1' },
    ...overrides
  };
}

describe('consumers/router', () => {
  it('routes by envelope.type to the correct handler', async () => {
    const handler = vi.fn().mockResolvedValue({ ok: true });
    const router = createEventRouter({ handlers: { 'snapshot.ready': handler } });

    const res = await router.route(envelope());

    expect(res).toEqual({ ok: true });
    expect(handler).toHaveBeenCalledWith({ snapshotId: 'snap-123', tenantId: 'tenant-1' });
  });

  it('returns failure when no handler is registered', async () => {
    const router = createEventRouter({ handlers: {} });

    const res = await router.route(envelope({ type: 'unknown' }));

    expect(res).toEqual({ ok: false, error: 'No handler for type unknown' });
  });

  it('handles handler exceptions as failure', async () => {
    const handler = vi.fn().mockRejectedValue(new Error('boom'));
    const router = createEventRouter({ handlers: { 'snapshot.ready': handler } });

    const res = await router.route(envelope());

    expect(res.ok).toBe(false);
    if (!res.ok) expect(res.error).toMatch(/boom/);
  });
});
