import { describe, it, expect, vi } from 'vitest';
import type { NotificationJob } from '../../consumers/types.js';
import { createFallbackChannel } from '../fallback.js';

function job(): NotificationJob {
  return {
    id: 'j1', tenantId: 't1', userId: 'a@b.com', channel: 'email', templateName: 'x', priority: 'normal', payload: {}, createdAt: new Date().toISOString(), attempt: 0, maxAttempts: 3
  };
}

describe('channels/fallback', () => {
  it('uses primary when success', async () => {
    const primary = { send: vi.fn().mockResolvedValue({ ok: true as const }) };
    const fallback = { send: vi.fn() };
    const ch = createFallbackChannel({ primary: primary as any, fallback: fallback as any });
    const res = await ch.send(job());
    expect(res).toEqual({ ok: true });
    expect(primary.send).toHaveBeenCalled();
    expect(fallback.send).not.toHaveBeenCalled();
  });

  it('uses fallback when primary fails', async () => {
    const primary = { send: vi.fn().mockResolvedValue({ ok: false as const, error: 'x' }) };
    const fallback = { send: vi.fn().mockResolvedValue({ ok: true as const }) };
    const ch = createFallbackChannel({ primary: primary as any, fallback: fallback as any });
    const res = await ch.send(job());
    expect(res).toEqual({ ok: true });
    expect(primary.send).toHaveBeenCalled();
    expect(fallback.send).toHaveBeenCalled();
  });

  it('returns failure when both fail', async () => {
    const primary = { send: vi.fn().mockResolvedValue({ ok: false as const, error: 'p' }) };
    const fallback = { send: vi.fn().mockResolvedValue({ ok: false as const, error: 'f' }) };
    const ch = createFallbackChannel({ primary: primary as any, fallback: fallback as any });
    const res = await ch.send(job());
    expect(res.ok).toBe(false);
  });
});
