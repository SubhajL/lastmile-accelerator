import { describe, it, expect, vi } from 'vitest';
import { createResendChannel } from '../resend.js';
import type { NotificationJob } from '../../consumers/types.js';

function job(overrides: Partial<NotificationJob> = {}): NotificationJob {
  return {
    id: 'j1',
    tenantId: 't1',
    userId: 'user@example.com',
    channel: 'email',
    templateName: 'publish-failed',
    priority: 'high',
    payload: { publishId: 'pb1' },
    createdAt: new Date().toISOString(),
    attempt: 0,
    maxAttempts: 3,
    ...overrides
  };
}

describe('channels/resend', () => {
  it('sends email via Resend API', async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 200, json: vi.fn() });
    const ch = createResendChannel({
      http: fetch as any,
      apiKey: 'key_123',
      from: 'noreply@example.com',
      resolveTo: async (j) => j.userId,
      renderTemplate: vi.fn().mockResolvedValue({ subject: 'sub', html: '<p>x</p>', text: 'x' })
    });

    const res = await ch.send(job());

    expect(res).toEqual({ ok: true });
    expect(fetch).toHaveBeenCalledWith(
      'https://api.resend.com/emails',
      expect.objectContaining({ method: 'POST' })
    );
  });

  it('returns failure on HTTP error', async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: false, status: 401, json: vi.fn() });
    const ch = createResendChannel({
      http: fetch as any,
      apiKey: 'badkey',
      from: 'noreply@example.com',
      resolveTo: async (j) => j.userId,
      renderTemplate: vi.fn().mockResolvedValue({ subject: 'sub', html: '<p>x</p>' })
    });

    const res = await ch.send(job());
    expect(res.ok).toBe(false);
  });
});
