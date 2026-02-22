import { describe, it, expect, vi } from 'vitest';
import { createTwilioSmsChannel } from './twilio';

function job(): any { return { templateName: 'snapshot-ready', payload: { id: 's1' }, userId: 'u1' }; }

describe('sms reliability', () => {
  it('timeout then retry succeeds', async () => {
    let calls = 0;
    const client = { messages: { create: vi.fn().mockImplementation(async () => {
      calls += 1;
      if (calls === 1) { await new Promise((r) => setTimeout(r, 50)); return {}; }
      return {};
    }) } } as any;
    const ch = createTwilioSmsChannel({ client, from: '+100000000', resolveTo: async () => '+12223334444', renderTemplate: async () => ({ text: 'hello' }), metrics: { increment: vi.fn() }, reliability: { timeoutMs: 5, retry: { max: 2, baseMs: 1, jitterPct: 0 } } });
    const res = await ch.send(job());
    expect(res.ok).toBe(true);
    expect(client.messages.create).toHaveBeenCalledTimes(2);
  });
});
