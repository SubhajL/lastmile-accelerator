import { describe, it, expect, vi } from 'vitest';
import { createEmailChannel } from './email';

function job(): any { return { templateName: 'fixes-applied', payload: { id: 'x' }, userId: 'u1' }; }

describe('email reliability', () => {
  it('sendMail timeout triggers retry then success', async () => {
    let calls = 0;
    const transporter = { sendMail: vi.fn().mockImplementation(async () => {
      calls += 1;
      if (calls === 1) { await new Promise((r) => setTimeout(r, 50)); return; }
      return;
    }) } as any;
    const renderTemplate = vi.fn().mockResolvedValue({ subject: 's', html: 'h', text: 't' });
    const ch = createEmailChannel({ transporter, from: 'noreply@example.com', renderTemplate, resolveTo: async () => 'a@b.com', reliability: { timeoutMs: 5, retry: { max: 2, baseMs: 1, jitterPct: 0 } } });
    const res = await ch.send(job());
    expect(res.ok).toBe(true);
    expect(transporter.sendMail).toHaveBeenCalledTimes(2);
  });

  it('non-retryable error fails fast, no retries', async () => {
    const transporter = { sendMail: vi.fn().mockRejectedValue(new Error('perm')) } as any;
    const renderTemplate = vi.fn().mockResolvedValue({ subject: 's', html: 'h', text: 't' });
    const ch = createEmailChannel({ transporter, from: 'noreply@example.com', renderTemplate, resolveTo: async () => 'a@b.com', reliability: { timeoutMs: 50, retry: { max: 3, baseMs: 1, jitterPct: 0 } } });
    const res = await ch.send(job());
    expect(res.ok).toBe(false);
    expect(transporter.sendMail).toHaveBeenCalledTimes(1);
  });
});
