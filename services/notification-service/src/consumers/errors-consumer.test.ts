import { describe, it, expect, vi } from 'vitest';
import type { CriticalErrorEvent } from './types.js';
import { createErrorsConsumer } from './errors-consumer.js';

function evt(overrides: Partial<CriticalErrorEvent> = {}): CriticalErrorEvent {
  return {
    errorId: 'e1',
    projectId: 'p1',
    tenantId: 't1',
    service: 'svc',
    message: 'boom',
    ...overrides
  };
}

describe('consumers/errors-consumer', () => {
  it('enqueues critical error notification with critical priority', async () => {
    const queue = { enqueue: vi.fn().mockResolvedValue('jid') };
    const c = createErrorsConsumer({ queue: queue as any, defaultMaxAttempts: 3, resolveRecipients: vi.fn().mockResolvedValue(['u1']) });

    const res = await c.handle(evt());

    expect(res).toEqual({ ok: true, enqueued: 1 });
    expect(queue.enqueue).toHaveBeenCalledWith(expect.objectContaining({ templateName: 'critical-error', priority: 'critical' }));
  });
});
