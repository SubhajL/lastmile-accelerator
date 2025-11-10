import { describe, it, expect, vi } from 'vitest';
import { createRateLimitedChannel } from '../../sms/limiter.js';

function makeJob(): any { return { id: 'j', channel: 'sms' }; }

describe('channels/sms/limiter', () => {
  it('allows within capacity and blocks exceeding', async () => {
    const inner = { send: vi.fn().mockResolvedValue({ ok: true as const }) };
    const limiter = { tryRemoveTokens: vi.fn().mockReturnValueOnce(true).mockReturnValueOnce(false) };

    const ch = createRateLimitedChannel({ inner: inner as any, limiter });

    expect(await ch.send(makeJob())).toEqual({ ok: true });
    const res2 = await ch.send(makeJob());
    expect(res2.ok).toBe(false);
    expect(limiter.tryRemoveTokens).toHaveBeenCalledTimes(2);
  });
});
