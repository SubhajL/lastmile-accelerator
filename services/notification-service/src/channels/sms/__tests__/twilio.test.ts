import { describe, it, expect, vi } from 'vitest';
import { createTwilioSmsChannel } from '../../sms/twilio.js';
import type { NotificationJob } from '../../../consumers/types.js';

function job(overrides: Partial<NotificationJob> = {}): NotificationJob {
  return {
    id: 'j1', tenantId: 't1', userId: '+15551234567', channel: 'sms', templateName: 'publish-failed', priority: 'high', payload: { snapshotId: 's1', error: 'boom' }, createdAt: new Date().toISOString(), attempt: 0, maxAttempts: 3,
    ...overrides
  };
}

describe('channels/sms/twilio', () => {
  it('sends SMS with Twilio client', async () => {
    const create = vi.fn().mockResolvedValue({ sid: 'SM123' });
    const client = { messages: { create } } as any;
    const render = vi.fn().mockResolvedValue({ subject: 'sub', text: 'Publish failed: boom', html: '<p>boom</p>' });

    const ch = createTwilioSmsChannel({ client, from: '+15550000000', resolveTo: async (j) => j.userId, renderTemplate: render, metrics: { increment: vi.fn() } });
    const res = await ch.send(job());

    expect(res).toEqual({ ok: true });
    expect(create).toHaveBeenCalledWith({ to: '+15551234567', from: '+15550000000', body: expect.stringContaining('Publish failed') });
  });

  it('returns failure on API error', async () => {
    const create = vi.fn().mockRejectedValue(new Error('Twilio error'));
    const client = { messages: { create } } as any;
    const render = vi.fn().mockResolvedValue({ subject: 'sub', text: 'x' });

    const ch = createTwilioSmsChannel({ client, from: '+15550000000', resolveTo: async (j) => j.userId, renderTemplate: render, metrics: { increment: vi.fn() } });
    const res = await ch.send(job());

    expect(res.ok).toBe(false);
  });
});
