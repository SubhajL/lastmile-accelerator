import { describe, it, expect, vi } from 'vitest';
import { createSmsFallbackToEmail } from '../../sms/fallback.js';

function makeJob(): any { return { id: 'j', channel: 'sms' }; }

describe('channels/sms/fallback', () => {
  it('uses email when sms fails', async () => {
    const sms = { send: vi.fn().mockResolvedValue({ ok: false as const, error: 'adapter_error' }) };
    const email = { send: vi.fn().mockResolvedValue({ ok: true as const }) };

    const ch = createSmsFallbackToEmail({ sms: sms as any, email: email as any });
    const res = await ch.send(makeJob());

    expect(res).toEqual({ ok: true });
    expect(sms.send).toHaveBeenCalled();
    expect(email.send).toHaveBeenCalled();
  });
});
