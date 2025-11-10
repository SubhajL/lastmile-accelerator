import { describe, it, expect, vi } from 'vitest';
import { createChannelRegistry, createDefaultChannelRegistry } from './registry.js';
import type { ChannelAdapter } from '../notifications/dispatcher.js';
import { createEmailChannel } from './email.js';

describe('channels/registry', () => {
  it('createChannelRegistry returns registered adapter by channel', () => {
    const email: ChannelAdapter = { send: vi.fn() } as any;
    const registry = createChannelRegistry({ email });

    expect(registry.get('email')).toBe(email);
    expect(registry.get('sms')).toBeUndefined();
  });

  it('createDefaultChannelRegistry wires email adapter correctly', async () => {
    const transporter = { sendMail: vi.fn().mockResolvedValue({}) };
    const renderTemplate = vi.fn().mockResolvedValue({ subject: 'Hello', html: '<p>Hi</p>', text: 'Hi' });
    const resolveTo = vi.fn().mockResolvedValue('user@example.com');

    const registry = createDefaultChannelRegistry({
      email: {
        transporter: transporter as any,
        renderTemplate,
        from: 'noreply@example.com',
        resolveTo
      }
    });

    const adapter = registry.get('email');
    expect(adapter).toBeDefined();

    const job: any = {
      id: 'j1', tenantId: 't1', userId: 'u1', channel: 'email', templateName: 'x',
      priority: 'normal', payload: {}, createdAt: new Date().toISOString(), attempt: 0, maxAttempts: 3
    };

    await adapter!.send(job);

    expect(renderTemplate).toHaveBeenCalled();
    expect(transporter.sendMail).toHaveBeenCalled();
  });
});
