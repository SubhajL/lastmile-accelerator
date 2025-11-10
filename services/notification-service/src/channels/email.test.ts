import { describe, it, expect, vi } from 'vitest';
import type { NotificationJob } from '../consumers/types.js';
import { createEmailChannel, type TemplateRenderer } from './email.js';

function makeJob(overrides: Partial<NotificationJob> = {}): NotificationJob {
  return {
    id: 'job-1',
    tenantId: 'tenant-1',
    userId: 'user-1',
    channel: 'email',
    templateName: 'snapshot-ready',
    priority: 'normal',
    payload: { snapshotId: 'snap-123' },
    createdAt: new Date().toISOString(),
    attempt: 0,
    maxAttempts: 3,
    ...overrides
  };
}

describe('channels/email', () => {
  it('send renders template and sends via transporter', async () => {
    const transporter = { sendMail: vi.fn().mockResolvedValue({ messageId: 'abc' }) };
    const render: TemplateRenderer = vi.fn().mockResolvedValue({
      subject: 'Snapshot ready',
      html: '<p>Ready</p>',
      text: 'Ready'
    });

    const channel = createEmailChannel({
      transporter: transporter as any,
      renderTemplate: render,
      from: 'noreply@example.com',
      resolveTo: vi.fn().mockResolvedValue('user@example.com')
    });

    const result = await channel.send(makeJob());

    expect(result).toEqual({ ok: true });
    expect(render).toHaveBeenCalledWith('snapshot-ready', { snapshotId: 'snap-123' });
    expect(transporter.sendMail).toHaveBeenCalledWith({
      from: 'noreply@example.com',
      to: 'user@example.com',
      subject: 'Snapshot ready',
      html: '<p>Ready</p>',
      text: 'Ready'
    });
  });

  it('send returns failure result when transporter throws', async () => {
    const transporter = { sendMail: vi.fn().mockRejectedValue(new Error('SMTP error')) };
    const render: TemplateRenderer = vi.fn().mockResolvedValue({ subject: 'x', html: 'y' });

    const channel = createEmailChannel({
      transporter: transporter as any,
      renderTemplate: render,
      from: 'noreply@example.com',
      resolveTo: vi.fn().mockResolvedValue('user@example.com')
    });

    const result = await channel.send(makeJob());

    expect(result.ok).toBe(false);
    if (result.ok === false) {
      expect(result.error).toMatch(/SMTP error/);
    }
  });
});
