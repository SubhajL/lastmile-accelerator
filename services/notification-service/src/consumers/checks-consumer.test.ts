import { describe, it, expect, vi } from 'vitest';
import type { CheckFailedEvent } from './types.js';
import { createChecksConsumer } from './checks-consumer.js';

function evt(overrides: Partial<CheckFailedEvent> = {}): CheckFailedEvent {
  return {
    checkId: 'c1',
    projectId: 'p1',
    tenantId: 't1',
    checkType: 'lint',
    severity: 'high',
    message: 'Failed',
    ...overrides
  };
}

describe('consumers/checks-consumer', () => {
  it('enqueues notification with priority by severity', async () => {
    const queue = { enqueue: vi.fn().mockResolvedValue('jid') };
    const c = createChecksConsumer({ queue: queue as any, defaultMaxAttempts: 3, resolveRecipients: vi.fn().mockResolvedValue(['u1']) });

    await c.handle(evt({ severity: 'low' }));
    expect(queue.enqueue).toHaveBeenCalledWith(expect.objectContaining({ templateName: 'check-failed', priority: 'normal' }));

    await c.handle(evt({ severity: 'critical' }));
    expect(queue.enqueue).toHaveBeenCalledWith(expect.objectContaining({ templateName: 'check-failed', priority: 'high' }));
  });
});
